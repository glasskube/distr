package agentmanifest

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"path"
	"text/template"

	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/glasskube/cloud/internal/types"
)

var (
	loginEndpoint     string
	resourcesEndpoint string
	statusEndpoint    string
)

func init() {
	if u, err := url.Parse(env.Host()); err != nil {
		panic(err)
	} else {
		u = u.JoinPath("api/v1/agent")
		loginEndpoint = u.JoinPath("login").String()
		resourcesEndpoint = u.JoinPath("resources").String()
		statusEndpoint = u.JoinPath("status").String()
	}
}

func Get(ctx context.Context, deploymentTarget types.DeploymentTarget, secret *string) (io.Reader, error) {
	if agentVersion, err := db.GetAgentVersionForDeploymentTargetID(ctx, deploymentTarget.ID); err != nil {
		return nil, err
	} else if tmpl, err := getTemplate(deploymentTarget, *agentVersion); err != nil {
		return nil, err
	} else {
		var buf bytes.Buffer
		return &buf, tmpl.Execute(&buf, getTemplateData(deploymentTarget, *agentVersion, secret))
	}
}

func getTemplateData(
	deploymentTarget types.DeploymentTarget,
	agentVersion types.AgentVersion,
	secret *string,
) map[string]any {
	return map[string]any{
		"loginEndpoint":     loginEndpoint,
		"resourcesEndpoint": resourcesEndpoint,
		"statusEndpoint":    statusEndpoint,
		"targetId":          deploymentTarget.ID,
		"targetSecret":      secret,
		"agentInterval":     env.AgentInterval(),
		"agentVersion":      agentVersion.Name,
	}
}

func getTemplate(deploymentTarget types.DeploymentTarget, agentVersion types.AgentVersion) (*template.Template, error) {
	if deploymentTarget.Type == types.DeploymentTypeDocker {
		return resources.GetTemplate(path.Join("agent/docker", agentVersion.ComposeFileRevision, "docker-compose.yaml"))
	} else {
		return resources.GetTemplate(path.Join("agent/kubernetes", agentVersion.ManifestFileRevision, "manifest.yaml.tmpl"))
	}
}
