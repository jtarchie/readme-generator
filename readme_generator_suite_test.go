//nolint:dupl
package main

import (
	"io"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func TestReadmeGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReadmeGenerator Suite")
}

var _ = Describe("CLI", func() {
	var server *ghttp.Server
	var readmeFile *os.File

	BeforeEach(func() {
		var err error
		readmeFile, err = os.CreateTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Run", func() {
		Context("Happy Path", func() {
			It("should run without errors", func() {
				// Setup
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"choices": []map[string]interface{}{
								{
									"message": map[string]interface{}{
										"content": "Generated README",
									},
								},
							},
						}),
					),
				)

				// Execute
				cli := CLI{
					// Initialize with appropriate values
					Glob:              "some glob pattern",
					Filename:          readmeFile.Name(),
					OpenAIAccessToken: "some token",
					BaseURL:           server.URL(), // Use the mock server URL
					Prompt:            "some prompt",
					Model:             "gpt-3.5-turbo",
				}
				err := cli.Run()

				// Assert
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Sad Path", func() {
			It("should return an error when the API call fails", func() {
				// Setup
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.RespondWith(http.StatusInternalServerError, "Internal Server Error"),
					),
				)

				// Execute
				cli := CLI{
					// Initialize with appropriate values
					Glob:              "some glob pattern",
					Filename:          readmeFile.Name(),
					OpenAIAccessToken: "some token",
					BaseURL:           server.URL(), // Use the mock server URL
					Prompt:            "some prompt",
					Model:             "gpt-3.5-turbo",
				}
				err := cli.Run()

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not translate"))
			})
		})
	})

	Describe("Run with different models", func() {
		It("should use the correct model for GPT-3", func() {
			// Setup
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/chat/completions"),
					func(w http.ResponseWriter, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						defer r.Body.Close()

						Expect(err).NotTo(HaveOccurred())
						Expect(string(body)).To(ContainSubstring(`gpt-3.5-turbo`))
					},
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
						"choices": []map[string]interface{}{
							{
								"message": map[string]interface{}{
									"content": "Generated README",
								},
							},
						},
					}),
				),
			)

			// Execute
			cli := CLI{
				Glob:              "some glob pattern",
				Filename:          readmeFile.Name(),
				OpenAIAccessToken: "some token",
				BaseURL:           server.URL(), // Use the mock server URL
				Prompt:            "some prompt",
				Model:             "gpt-3.5-turbo",
			}
			err := cli.Run()

			// Assert
			Expect(err).NotTo(HaveOccurred())
		})

		It("should use the correct model for GPT-4", func() {
			// Setup
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/chat/completions"),
					func(w http.ResponseWriter, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						defer r.Body.Close()

						Expect(err).NotTo(HaveOccurred())
						Expect(string(body)).To(ContainSubstring(`gpt-4`))
					},
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
						"choices": []map[string]interface{}{
							{
								"message": map[string]interface{}{
									"content": "Generated README",
								},
							},
						},
					}),
				),
			)

			cli := CLI{
				Glob:              "some glob pattern",
				Filename:          readmeFile.Name(),
				OpenAIAccessToken: "some token",
				BaseURL:           server.URL(), // Use the mock server URL
				Prompt:            "some prompt",
				Model:             "gpt-4",
			}
			err := cli.Run()

			// Assert
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
