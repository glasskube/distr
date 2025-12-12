package agentconnect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/glasskube/distr/internal/customdomains"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

func BuildConnectURL(targetID uuid.UUID, org types.Organization, targetSecret string) (string, error) {
	u, err := url.Parse(customdomains.AppDomainOrDefault(org))
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("targetId", targetID.String())
	query.Set("targetSecret", targetSecret)
	u = u.JoinPath("/api/v1/connect")
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func BuildPreConnectURL(targetID uuid.UUID, org types.Organization, targetSecret string) (string, error) {
	u, err := url.Parse(customdomains.AppDomainOrDefault(org))
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("targetId", targetID.String())
	query.Set("targetSecret", targetSecret)
	u = u.JoinPath("/api/v1/pre-connect")
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func GenerateConnectScript(targetID uuid.UUID, org types.Organization, targetSecret string) (string, error) {
	connectURL, err := BuildConnectURL(targetID, org, targetSecret)
	if err != nil {
		return "", fmt.Errorf("failed to build connect URL: %w", err)
	}

	var script strings.Builder
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -euo pipefail\n\n")

	if org.PreConnectScript != nil && strings.TrimSpace(*org.PreConnectScript) != "" {
		script.WriteString("# Pre-connect script\n")
		script.WriteString(*org.PreConnectScript)
		script.WriteString("\n\n")
	}

	script.WriteString("# Connect to Distr agent\n")
	script.WriteString(fmt.Sprintf("curl -fsSL '%s'\n", connectURL))

	if org.PostConnectScript != nil && strings.TrimSpace(*org.PostConnectScript) != "" {
		script.WriteString("\n\n# Post-connect script\n")
		script.WriteString(*org.PostConnectScript)
		script.WriteString("\n")
	}

	return script.String(), nil
}

func GenerateConnectCommand(
	deploymentTarget types.DeploymentTarget,
	org types.Organization,
	targetSecret string,
) (string, error) {
	preConnectURL, err := BuildPreConnectURL(deploymentTarget.ID, org, targetSecret)
	if err != nil {
		return "", fmt.Errorf("failed to build pre-connect URL: %w", err)
	}

	var command strings.Builder
	command.WriteString("bash <(curl -fsSL '")
	command.WriteString(preConnectURL)
	command.WriteString("')")

	return command.String(), nil
}
