package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/kong"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sashabaranov/go-openai"
)

type CLI struct {
	Glob              string `required:"" help:"glob pattern of files to read to help determine the README content"`
	Filename          string `required:"" default:"README.md" help:"name of the file to output the generated readme"`
	OpenAIAccessToken string `help:"the API token for the OpenAI API" required:"" env:"OPENAI_ACCESS_TOKEN"`
	BaseURL           string `help:"url of the OpenAI HTTP domain" default:"https://api.openai.com/v1"`
	Prompt            string `required:"" help:"additional prompt information when generating the README"`
	Model             string `required:"" default:"gpt-3.5-turbo" help:"the model to use for the OpenAI API" enum:"gpt-3.5-turbo,gpt-4"`
}

func main() {
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

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			Content: heredoc.Doc(`
				Over the next few prompts from the user, you will receive the contents of several files.
				Please take the input of all the files without returning any prose, just confirm receipt and waiting for the next file or prompt.
				The format of the file with be two headers:
				- filename: this contains the name of the file
				- contents: the first bunch of content from the file

				When the user finally provides a prompt, which is not file, please do you best to follow that prompt.
			`),
		},
	}

	for _, match := range matches {
		filename := filepath.Join(absPath, match)

		fmt.Printf("filename: %s\n", filename)

		contents, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("could not read file: %w", err)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			Content: strings.TrimSpace(fmt.Sprintf(`
				filename: %s
				contents: %s
			`,
				strings.Replace(match, basepath, "", 1),
				string(contents)[0:min(4000, len(contents))],
			)),
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		Content: heredoc.Doc(fmt.Sprintf(`
			Given all the files above, please write a README file for this code.
			Ensure that the copy is in active voice, removes any duplication, and
			is approachable to all software engineers.

			It must include the following:
			- Name of the project
			- A brief the description of the intention of the codebase.
			- A list short list of high level feature set.
			- Usage example of how to install the library
				- if it is a CLI, show an example invocation with description all the parameters
				- if it library to be invoked in code, please give an example of how to use it. If
				  there multiple functions give three of the most obvious starting points
			- Any other sections that would be useful to a README. If
			  it requires additional changes from the author insert text that "FIXME" with
				a description that they should do.

			The following is additional prompting by the author:

			%s
		`, c.Prompt)),
	})

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
		return fmt.Errorf("could not translate: %w", err)
	}

	readme := response.Choices[0].Message.Content

	err = os.WriteFile(c.Filename, []byte(readme), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write file %q: %w", c.Filename, err)
	}

	return nil
}
