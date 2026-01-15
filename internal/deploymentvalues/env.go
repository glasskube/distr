package deploymentvalues

import (
	"fmt"

	"github.com/distr-sh/distr/internal/types"
)

type EnvFileDataAccessor interface {
	GetEnvFileData() []byte
}

func EnvFileReplaceSecrets(d EnvFileDataAccessor, secrets []types.SecretWithUpdatedBy) ([]byte, error) {
	if data := d.GetEnvFileData(); data == nil {
		return nil, nil
	} else if tpl, err := parseTemplateBytes("envFile", data); err != nil {
		return nil, fmt.Errorf("deployment env file template parsing error: %w", err)
	} else if data, err := executeTemplate(tpl, getTemplateData(secrets)); err != nil {
		return nil, fmt.Errorf("deployment env file template execution error: %w", err)
	} else {
		return data, nil
	}
}
