//nolint:gosec
package main

const setupAssistant = `
Over the next few prompts from the user, you will receive the contents of several files.
Please take the input of all the files without returning any prose, just confirm receipt and waiting for the next file or prompt.
The format of the file with be two headers:
- filename: this contains the name of the file
- contents: the first bunch of content from the file

When the user finally provides a prompt, which is not file, please do you best to follow that prompt.
`

const readmePrompt = `
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
`
