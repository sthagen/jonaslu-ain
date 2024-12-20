package disk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/pkg/errors"
)

var backendPrioOrder = []string{"curl", "httpie", "wget"}

var starterTemplate = `[Host]
http://localhost:${PORT}

[Headers]
Content-Type: application/json

# [Query]
# id=1

# [Method]
# POST

# [Body]
# {
#    "some": "json"
# }

[Config]
Timeout=3
# queryDelim=&

[Backend]
{{Backends}}

[BackendOptions]
{{BackendOptions}}

# Short help:
# Comments start with hash-sign (#) and are ignored
# ${VARIABLES} are replaced with the .env-file or environment variable value
# $(executables.sh) are replaced with the output of the executable
# More in depth help here: https://github.com/jonaslu/ain`

func getUsefulBackendOptions(backendBinary string) string {
	switch backendBinary {
	case "curl":
		return "-sS # Makes curl suppress its progress bar in a pipe"
	case "wget":
		return "-q # Makes wget suppress its progress output"
	case "http":
		return "-b # Makes httpie output only the body and not the headers"
	}

	return ""
}

func getPresentBackendBinaries() ([]string, []string) {
	presentBackends := []string{}
	usefulBackendOptions := []string{}

	for _, backendTemplateName := range backendPrioOrder {
		backendConstructor := call.ValidBackends[backendTemplateName]
		if _, err := exec.LookPath(backendConstructor.BinaryName); err == nil {
			presentBackends = append(presentBackends, backendTemplateName)

			if backendOptions := getUsefulBackendOptions(backendConstructor.BinaryName); backendOptions != "" {
				usefulBackendOptions = append(usefulBackendOptions, getUsefulBackendOptions(backendConstructor.BinaryName))
			}
		}
	}

	return presentBackends, usefulBackendOptions
}

func GenerateEmptyTemplates(templateFileNames []string) error {
	presentBackends, usefulBackendOptions := getPresentBackendBinaries()
	if len(presentBackends) == 0 {
		presentBackends = []string{`# No backend binaries found, please install at least one.
# See here for more help: https://github.com/jonaslu/ain#pre-requisites`}
	} else {
		for i := 1; i < len(presentBackends); i++ {
			presentBackends[i] = "# " + presentBackends[i]
		}

		for i := 1; i < len(usefulBackendOptions); i++ {
			usefulBackendOptions[i] = "# " + usefulBackendOptions[i]
		}
	}

	// text/template is too complicated for this, we're replacing strings until it feels too heavy
	starterTemplate = strings.ReplaceAll(starterTemplate, "{{Backends}}", strings.Join(presentBackends, "\n"))

	if len(usefulBackendOptions) > 0 {
		starterTemplate = strings.ReplaceAll(starterTemplate, "{{BackendOptions}}", strings.Join(usefulBackendOptions, "\n"))
	} else {
		starterTemplate = strings.ReplaceAll(starterTemplate, "[BackendOptions]\n", "")
		starterTemplate = strings.ReplaceAll(starterTemplate, "{{BackendOptions}}\n\n", "")
	}

	if len(templateFileNames) == 0 {
		_, err := fmt.Fprintln(os.Stdout, starterTemplate)
		return err
	}

	for _, filename := range templateFileNames {
		_, err := os.Stat(filename)

		if !os.IsNotExist(err) {
			return errors.Errorf("cannot write basic template. File already exists %s", filename)
		}

		err = os.WriteFile(filename, []byte(starterTemplate), 0644)

		if err != nil {
			return errors.Wrapf(err, "could not write basic template to file %s", filename)
		}
	}

	return nil
}
