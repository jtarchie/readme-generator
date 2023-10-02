//nolint:gosec
package main

const setupAssistant = `
You will receive the following about a file.
- filename: this contains the name of the file
- contents: the first bunch of content from the file

Please provide context of the file with the following criteria:
- Feature set the file provides
- Brief one sentence summary of the file.
- A good code example of the file being used.
`

const readmePrompt = `
Given all the files above, please write a README file for this code.
Ensure that the copy is in active voice, removes any duplication, and
is approachable to all software engineers.

Please include the following, if cannot be inferred, please leave it out:
- Name of the project
- A brief the description of the intention of the codebase.
- A list short list of high level feature set.
- Usage example of how to install the library
	- if it is a CLI, show an example invocation with description all the parameters
	- if it library to be invoked in code, please give an example of how to use it. If
		there multiple functions give three of the most obvious starting points
	- if it is a website, an example of how to create a post
- Any other sections that would be useful to a README. If
	it requires additional changes from the author insert text that "FIXME" with
	a description that they should do.

The following is additional prompting by the author:

%s
`
