package deploymentvalues

import (
	"bytes"
	"text/template"

	"github.com/glasskube/distr/internal/types"
)

type templateData struct {
	Secrets map[string]string
}

func getTemplateData(secrets []types.SecretWithUpdatedBy) templateData {
	data := templateData{
		Secrets: make(map[string]string),
	}
	for _, secret := range secrets {
		data.Secrets[secret.Key] = secret.Value
	}
	return data
}

func parseTemplateBytes(name string, data []byte) (*template.Template, error) {
	return template.New(name).Option("missingkey=error").Parse(string(data))
}

func executeTemplate(tpl *template.Template, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}
