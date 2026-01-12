package deploymentvalues

import (
	"fmt"

	"github.com/glasskube/distr/internal/types"
	"gopkg.in/yaml.v3"
)

type ValuesYAMLAccessor interface {
	GetValuesYAML() []byte
}

func ParsedValuesFileReplaceSecrets(d ValuesYAMLAccessor, secrets []types.SecretWithUpdatedBy) (map[string]any, error) {
	if data := d.GetValuesYAML(); data == nil {
		return nil, nil
	} else if tpl, err := parseTemplateBytes("valuesYaml", data); err != nil {
		return nil, fmt.Errorf("deployment values file template parsing error: %w", err)
	} else if data, err := executeTemplate(tpl, getTemplateData(secrets)); err != nil {
		return nil, fmt.Errorf("deployment values file template execution error: %w", err)
	} else {
		return parseDeploymentValuesYAML(data)
	}
}

func ParsedValuesFile(d ValuesYAMLAccessor) (result map[string]any, err error) {
	if data := d.GetValuesYAML(); data == nil {
		return nil, nil
	} else {
		return parseDeploymentValuesYAML(data)
	}
}

func parseDeploymentValuesYAML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("cannot parse Deployment values file: %w", err)
	}
	return result, nil
}
