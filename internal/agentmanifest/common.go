package agentmanifest

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"path"
	"text/template"

	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/glasskube/cloud/internal/types"
)

var (
	loginEndpoint     string
	manifestEndpoint  string
	resourcesEndpoint string
	statusEndpoint    string
)

func init() {
	if u, err := url.Parse(env.Host()); err != nil {
		panic(err)
	} else {
		u = u.JoinPath("api/v1/agent")
		loginEndpoint = u.JoinPath("login").String()
		manifestEndpoint = u.JoinPath("manifest").String()
		resourcesEndpoint = u.JoinPath("resources").String()
		statusEndpoint = u.JoinPath("status").String()
	}
}

func Get(ctx context.Context, deploymentTarget types.DeploymentTargetWithCreatedBy, secret *string) (io.Reader, error) {
	if tmpl, err := getTemplate(deploymentTarget); err != nil {
		return nil, err
	} else {
		var buf bytes.Buffer
		return &buf, tmpl.Execute(&buf, getTemplateData(deploymentTarget, secret))
	}
}

func getTemplateData(
	deploymentTarget types.DeploymentTargetWithCreatedBy,
	secret *string,
) map[string]any {
	return map[string]any{
		"agentInterval":     env.AgentInterval(),
		"agentVersion":      deploymentTarget.AgentVersion.Name,
		"agentVersionId":    deploymentTarget.AgentVersion.ID,
		"targetId":          deploymentTarget.ID,
		"targetSecret":      secret,
		"loginEndpoint":     loginEndpoint,
		"manifestEndpoint":  manifestEndpoint,
		"resourcesEndpoint": resourcesEndpoint,
		"statusEndpoint":    statusEndpoint,
	}
}

func getTemplate(deploymentTarget types.DeploymentTargetWithCreatedBy) (*template.Template, error) {
	if deploymentTarget.Type == types.DeploymentTypeDocker {
		return resources.GetTemplate(path.Join(
			"agent/docker",
			deploymentTarget.AgentVersion.ComposeFileRevision,
			"docker-compose.yaml",
		))
	} else {
		return resources.GetTemplate(path.Join(
			"agent/kubernetes",
			deploymentTarget.AgentVersion.ManifestFileRevision,
			"manifest.yaml.tmpl",
		))
	}
}
