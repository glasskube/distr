package main

import (
	"context"
	"github.com/glasskube/distr/api"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func startMetrics(ctx context.Context) {
	// TODO only if metrics enabled

	var cpuCapacitySum int64
	var cpuRequestSum int64
	var memoryCapacitySum int64
	var memoryRequestSum int64
	if nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
		logger.Error("getting nodes failed", zap.Error(err))
	} else {
		for _, node := range nodes.Items {
			logger.Info("node", zap.String("name", node.Name))
			cpuCapacitySum = cpuCapacitySum + node.Status.Capacity.Cpu().MilliValue()
			memoryCapacitySum = memoryCapacitySum + node.Status.Capacity.Memory().Value()

			/*if pods, err := getNodePods(ctx, node); err != nil {
				logger.Error("failed to get pods of node", zap.String("node", node.Name), zap.Error(err))
			} else {
				for _, pod := range pods.Items {
					if pod.Spec.Resources != nil && pod.Spec.Resources.Requests.Cpu() != nil {
						cpuRequestSum = cpuRequestSum + pod.Spec.Resources.Requests.Cpu().MilliValue()
					} else {
						logger.Debug("pod has no cpu request", zap.String("pod", pod.Name))
					}
					if pod.Spec.Resources != nil && pod.Spec.Resources.Requests.Memory() != nil {
						memoryRequestSum = memoryRequestSum + pod.Spec.Resources.Requests.Memory().Value()
					} else {
						logger.Debug("pod has no memory request", zap.String("pod", pod.Name))
					}
				}
			}*/

			// TODO this instead???
			if nodeMetrics, err := metricsClientSet.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{}); err != nil {
				logger.Error("getting node metrics failed", zap.Error(err))
			} else {
				logger.Debug("pod metrics", zap.Any("cpuUsage", nodeMetrics.Usage.Cpu().MilliValue()),
					zap.Any("memUsage", nodeMetrics.Usage.Memory().Value()))
				cpuRequestSum = cpuRequestSum + nodeMetrics.Usage.Cpu().MilliValue()
				memoryRequestSum = memoryRequestSum + nodeMetrics.Usage.Memory().Value()
			}
		}
	}

	logger.Debug("pod metrics", zap.Any("cpuUsageSum", cpuRequestSum),
		zap.Any("memUsageSum", memoryRequestSum))

	if cpuCapacitySum > 0 && memoryCapacitySum > 0 {
		if err := agentClient.ReportMetrics(ctx, api.AgentDeploymentTargetMetrics{
			CPUCoresM:   cpuCapacitySum,
			CPUUsage:    float64(cpuRequestSum) / float64(cpuCapacitySum),
			MemoryBytes: memoryCapacitySum,
			MemoryUsage: float64(memoryRequestSum) / float64(memoryCapacitySum),
		}); err != nil {
			logger.Error("failed to report metrics", zap.Error(err))
		}
	}
}
