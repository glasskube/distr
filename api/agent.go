package api

type KubernetesAgentResource struct {
	Namespace  string                     `json:"namespace"`
	Deployment *KubernetesAgentDeployment `json:"deployment"`
}

type KubernetesAgentDeployment struct {
	RevisionID   string         `json:"revisionId"`
	ReleaseName  string         `json:"releaseName"`
	ChartUrl     string         `json:"chartUrl"`
	ChartName    string         `json:"chartName"`
	ChartVersion string         `json:"chartVersion"`
	Values       map[string]any `json:"values"`
}
