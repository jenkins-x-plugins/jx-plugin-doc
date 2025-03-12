package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	samplePage = `
## jx-gitops annotate

Annotates all kubernetes resources in the given directory tree

### Usage

    jx-gitops annotate

### Synopsis

Annotates all kubernetes resources in the given directory tree

### Examples

  # updates recursively annotates all resources in the current directory
  jx-gitops annotate myannotate=cheese another=thing
  # updates recursively all resources
  jx-gitops annotate --dir myresource-dir foo=bar

### Options
`
)

func TestWrapExamplesInCodeBlock(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "### Examples\n\n  # upgrades the plugin binaries\n  jx upgrade\n\n### Options\n",
			expected: "### Examples\n\n  ```bash\n  # upgrades the plugin binaries\n  jx upgrade\n\n  ```\n### Options\n",
		},

		{
			input:    "### Synopsis\n\nInstalls the git operator in a cluster\n\n### Examples\n\n  * installs the git operator from inside a git clone and prompt for the user/token if required\n  \n  ```bash\n  jx-admin operator\n  ```\n  \n  * installs the git operator from inside a git clone specifying the user/token\n  \n  ```bash\n  jx-admin operator --username mygituser --token mygittoken\n  ```\n  \n  * installs the git operator with the given git clone URL\n  \n  ```bash\n  jx-admin operator --url https://github.com/myorg/environment-mycluster-dev.git --username myuser --token myuser\n  ```\n  \n  * display what helm command will install the git operator\n  \n  ```bash\n  jx-admin operator --dry-run\n  ```\n\n### Options",
			expected: "### Synopsis\n\nInstalls the git operator in a cluster\n\n### Examples\n\n  * installs the git operator from inside a git clone and prompt for the user/token if required\n  \n  ```bash\n  jx-admin operator\n  ```\n  \n  * installs the git operator from inside a git clone specifying the user/token\n  \n  ```bash\n  jx-admin operator --username mygituser --token mygittoken\n  ```\n  \n  * installs the git operator with the given git clone URL\n  \n  ```bash\n  jx-admin operator --url https://github.com/myorg/environment-mycluster-dev.git --username myuser --token myuser\n  ```\n  \n  * display what helm command will install the git operator\n  \n  ```bash\n  jx-admin operator --dry-run\n  ```\n\n### Options",
		},
	}

	for _, tc := range testCases {
		got := WrapExamplesInCodeBlock(tc.input)
		assert.Equal(t, tc.expected, got, "for input %s", tc.input)
	}
}

func TestReadCobraDescription(t *testing.T) {
	desc := ReadCobraDescription(samplePage)
	assert.Equal(t, "Annotates all kubernetes resources in the given directory tree", desc, "failed to find description")
}
