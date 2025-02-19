package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/glasskube/distr/api"
	"github.com/google/uuid"
)

type AgentDeployment struct {
	ID          uuid.UUID `json:"id"`
	RevisionID  uuid.UUID `json:"revisionId"`
	ProjectName string    `json:"projectName"`
}

func (d *AgentDeployment) FileName() string {
	return path.Join(agentDeploymentDir(), d.ID.String())
}

func agentDeploymentDir() string {
	return path.Join(ScratchDir(), "deployments")
}

func NewAgentDeployment(deployment api.DockerAgentDeployment) (*AgentDeployment, error) {
	if name, err := getProjectName(deployment.ComposeFile); err != nil {
		return nil, err
	} else {
		return &AgentDeployment{ID: deployment.ID, RevisionID: deployment.RevisionID, ProjectName: name}, nil
	}
}

func getProjectName(data []byte) (string, error) {
	if compose, err := DecodeComposeFile(data); err != nil {
		return "", err
	} else if name, ok := compose["name"].(string); !ok {
		return "", fmt.Errorf("name is not a string")
	} else {
		return name, nil
	}
}

func GetExistingDeployments() ([]AgentDeployment, error) {
	if entries, err := os.ReadDir(agentDeploymentDir()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	} else {
		result := make([]AgentDeployment, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				if file, err := os.Open(path.Join(agentDeploymentDir(), entry.Name())); err != nil {
					return nil, err
				} else {
					defer file.Close()
					var d AgentDeployment
					if err := json.NewDecoder(file).Decode(&d); err != nil {
						return nil, err
					}
					result = append(result, d)
				}
			}
		}
		return result, nil
	}
}

func SaveDeployment(deployment AgentDeployment) error {
	if err := os.MkdirAll(path.Dir(deployment.FileName()), 0o700); err != nil {
		return err
	}

	file, err := os.Create(deployment.FileName())
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(deployment); err != nil {
		return err
	}

	return nil
}

func DeleteDeployment(deployment AgentDeployment) error {
	return os.Remove(deployment.FileName())
}
