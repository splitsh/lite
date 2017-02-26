package splitter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

const (
	schema = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "Project configuration",
    "type" : "object",
    "additionalProperties": false,
    "required": ["git-version", "subtrees"],
    "properties": {
        "git-version": { "type": "string" },
        "subtrees": {
            "type": "object",
            "patternProperties": {
                "^\\w[\\w\\d\\-_\\.]+$": {
                    "title": "Split",
                    "type": "object",
                    "required": ["target", "prefixes"],
                    "additionalProperties": false,
                    "properties": {
                        "prefixes": {
                            "type": "array",
                            "items": { "type": "string" }
                        },
                        "target": { "type": "string" }
                    }
                }
            }
        }
    }
}
`
)

// Project represents a project
type Project struct {
	Subtrees   map[string]*Subtree `json:"subtrees"`
	GitVersion string              `json:"git-version"`
}

// Subtree represents a split configuration
type Subtree struct {
	Prefixes []string `json:"prefixes"`
	Target   string
}

// NewProject creates a project from a JSON string
func NewProject(config []byte) (*Project, error) {
	if err := validateProject(config); err != nil {
		return nil, err
	}

	var project *Project
	if err := json.Unmarshal(config, &project); err != nil {
		return nil, err
	}

	return project, nil
}

// validateProject validates a project JSON string
func validateProject(config []byte) error {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	if err != nil {
		return fmt.Errorf("Could not create the JSON validator: %s", err)
	}

	result, err := schema.Validate(gojsonschema.NewStringLoader(string(config)))
	if err != nil {
		return fmt.Errorf("Project configuration is not valid JSON: %s", err)
	}

	if !result.Valid() {
		errors := []string{}
		for _, desc := range result.Errors() {
			errors = append(errors, desc.Description())
		}

		return fmt.Errorf("Project configuration is not valid JSON: %s", strings.Join(errors, ", "))
	}

	return nil
}
