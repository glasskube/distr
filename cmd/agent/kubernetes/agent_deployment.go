package main

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

const LabelDeplyoment = "agent.glasskube.cloud/deployment"

type AgentDeployment struct {
	ReleaseName  string `json:"releaseName"`
	RevisionID   string `json:"revisionId"`
	HelmRevision int    `json:"helmRevision"`
}

func (d *AgentDeployment) SecretName() string {
	return fmt.Sprintf("cloud.glasskube.agent.v1.%v", d.ReleaseName)
}

func GetExistingDeployments(ctx context.Context, namespace string) ([]AgentDeployment, error) {
	if secrets, err := k8sClient.CoreV1().Secrets(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: LabelDeplyoment}); err != nil {
		return nil, err
	} else {
		deployments := make([]AgentDeployment, len(secrets.Items))
		for i, secret := range secrets.Items {
			var deployment AgentDeployment
			if err := json.Unmarshal(secret.Data["release"], &deployment); err != nil {
				return nil, err
			} else {
				deployments[i] = deployment
			}
		}
		return deployments, nil
	}
}

func SaveDeployment(ctx context.Context, namespace string, deployment AgentDeployment) error {
	cfg := applyconfigurationscorev1.Secret(deployment.SecretName(), namespace)
	cfg.WithLabels(map[string]string{LabelDeplyoment: deployment.ReleaseName})
	if data, err := json.Marshal(deployment); err != nil {
		return err
	} else {
		cfg.WithData(map[string][]byte{"release": data})
	}
	_, err := k8sClient.CoreV1().Secrets(namespace).Apply(
		ctx,
		cfg,
		metav1.ApplyOptions{Force: true, FieldManager: "glasskube-cloud-agent"},
	)
	return err
}
