package sections

import (
	"errors"
	"strings"

	"github.com/jonaslu/ain/data"
)

func ParseHostSection(template data.Template, templateSections *data.TemplateSections) (data.Warnings, error) {
	warnings := data.Warnings{}

	var startingHostLine data.SourceMarker
	var hostLines data.Template
	var capturing = -1

	for _, templateLine := range template {
		lineContents := templateLine.LineContents

		if lineContents == "[Host]" {
			if len(hostLines) > 0 {
				warning := data.Warning{
					Message:      "Several [Host] sections found",
					TemplateLine: templateLine,
				}

				warnings = append(warnings, warning)
			}

			startingHostLine = templateLine
			capturing = 0
			continue
		}

		if strings.HasPrefix(lineContents, "[") {
			if capturing == 0 {
				warning := data.Warning{
					Message:      "Empty [Host] section",
					TemplateLine: startingHostLine,
				}

				warnings = append(warnings, warning)
			}

			capturing = -1
			continue
		}

		if capturing >= 0 {
			hostLines = append(hostLines, templateLine)
			capturing++
		}
	}

	if len(hostLines) == 0 {
		return warnings, errors.New("No mandatory [Host] section found")
	}

	// Set the URL
	if len(hostLines) > 1 {

		for _, hostLine := range hostLines {
			warning := data.Warning{
				Message:      "Found serveral host lines",
				TemplateLine: hostLine,
			}

			warnings = append(warnings, warning)
		}
	}

	// !! TODO !! Parse URL as valid too
	templateSections.Host = hostLines[len(hostLines)-1].LineContents

	return warnings, nil
}
