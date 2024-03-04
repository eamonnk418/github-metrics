package dependabot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"
)

func GenerateDependabotConfig(path string) (*bytes.Buffer, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading template file: %w", err)
	}

	fmt.Println("Template Content:", string(content))

	dependabotYml, err := template.New("dependabot").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := dependabotYml.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	return &buf, nil
}
