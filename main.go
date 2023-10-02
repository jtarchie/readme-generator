package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/kong"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sashabaranov/go-openai"
)

const maxTokens = 4000

type CLI struct {
	Glob              string `help:"glob pattern of files to read to help determine the README content" required:""`
	Filename          string `default:"README.md"                                                       help:"name of the file to output the generated readme" required:""`
	OpenAIAccessToken string `env:"OPENAI_ACCESS_TOKEN"                                                 help:"the API token for the OpenAI API"                required:""`
	BaseURL           string `default:"https://api.openai.com/v1"                                       help:"url of the OpenAI HTTP domain"`
	Context           string `help:"additional context information when generating the README"          required:""`
	Model             string `default:"gpt-3.5-turbo"                                                   enum:"gpt-3.5-turbo,gpt-4"                             help:"the model to use for the OpenAI API" required:""`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := CLI{}
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (c *CLI) Run() error {
	basepath, pattern := doublestar.SplitPattern(c.Glob)

	absPath, err := filepath.Abs(basepath)
	if err != nil {
		return fmt.Errorf("could not resolve absolute path: %w", err)
	}

	matches, err := doublestar.Glob(os.DirFS(absPath), pattern)
	if err != nil {
		return fmt.Errorf("could not matches for %q: %w", c.Glob, err)
	}

	summaries := []string{}

	for _, match := range matches {
		filename := filepath.Join(absPath, match)

		slog.Info("processing file", slog.String("filename", filename))

		contents, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: heredoc.Doc(setupAssistant),
			},
			{
				Role: openai.ChatMessageRoleUser,
				Content: strings.TrimSpace(fmt.Sprintf(`
				filename: %s
				contents: %s
			`,
					strings.Replace(match, basepath, "", 1),
					string(contents)[0:min(maxTokens, len(contents))],
				)),
			},
		}

		summary, err := c.runPrompt(messages)
		if err != nil {
			return fmt.Errorf("could not process OpenAI for file %q: %w", match, err)
		}

		summaries = append(summaries, summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return len(summaries[i]) > len(summaries[j])
	})

	fullSummary := strings.Join(summaries, "\n---\n")
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: heredoc.Doc(fmt.Sprintf(readmePrompt, c.Context)),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fullSummary[0:min(len(fullSummary), maxTokens)],
		},
	}

	slog.Info("processing readme",
		slog.Int("files", len(summaries)),
	)

	readme, err := c.runPrompt(messages)
	if err != nil {
		return fmt.Errorf("could not process OpenAI for summary: %w", err)
	}

	err = os.WriteFile(c.Filename, []byte(readme), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write file %q: %w", c.Filename, err)
	}

	return nil
}

func (c *CLI) runPrompt(messages []openai.ChatCompletionMessage) (string, error) {
	config := openai.DefaultConfig(c.OpenAIAccessToken)
	config.BaseURL = c.BaseURL
	client := openai.NewClientWithConfig(config)

	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.Model,
			Messages: messages,
		})
	if err != nil {
		return "", fmt.Errorf("could not translate: %w", err)
	}

	return response.Choices[0].Message.Content, nil
}
