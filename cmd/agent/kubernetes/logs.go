package main

import (
	"bufio"
	"context"
	"time"

	"github.com/glasskube/distr/internal/agentlogs"
	"github.com/google/uuid"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

type logsWatcher struct {
	logsExporter agentlogs.Exporter
	last         map[uuid.UUID]metav1.Time
	namespace    string
}

func NewLogsWatcher(namespace string) *logsWatcher {
	return &logsWatcher{
		logsExporter: agentlogs.ChunkExporter(agentClient, 100),
		last:         make(map[uuid.UUID]metav1.Time),
		namespace:    namespace,
	}
}

func (lw *logsWatcher) Watch(ctx context.Context, d time.Duration) {
	tick := time.Tick(d)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			lw.collect(ctx)
		}
	}
}

func (lw *logsWatcher) collect(ctx context.Context) {
	existingDeployments, err := GetExistingDeployments(ctx, lw.namespace)
	if err != nil {
		logger.Error("could not get existing deployments", zap.Error(err))
		return
	}

	collector := agentlogs.NewCollector()

	for _, d := range existingDeployments {
		if !d.LogsEnabled {
			continue
		}

		var resources []runtime.Object
		if resUnstr, err := GetHelmManifest(ctx, lw.namespace, d.ReleaseName); err != nil {
			logger.Error("could not get existing deployments", zap.Error(err))
			continue
		} else {
			resources = FromUnstructuredSlice(resUnstr)
		}

		deploymentCollector := collector.For(d)
		now := metav1.Now()
		var toplevelErr error

	resourcesLoop:
		for _, obj := range resources {
			logger := logger.With(zap.String("resource", obj.GetObjectKind().GroupVersionKind().Kind))

			logOptions := corev1.PodLogOptions{Timestamps: true}
			if since, ok := lw.last[d.ID]; ok {
				logOptions.SinceTime = &since
			} else {
				logOptions.SinceTime = &now
			}

			responseMap, err :=
				polymorphichelpers.AllPodLogsForObjectFn(k8sConfigFlags, obj, &logOptions, 10*time.Second, true)
			if err != nil {
				// not being able to get logs for all resource types is normal so we only want to call abort when an
				// API error is encountered.
				if _, ok := err.(errors.APIStatus); ok {
					logger.Warn("could not get logs", zap.Error(err))
					toplevelErr = err
					break
				} else {
					logger.Debug("could not get logs", zap.Error(err))
				}
			}

			for ref, resp := range responseMap {
				err := func() error {
					rc, err := resp.Stream(ctx)
					if err != nil {
						logger.Warn("could not get logs for pod", zap.Error(err))
						return err
					}
					defer rc.Close()
					sc := bufio.NewScanner(rc)
					for sc.Scan() {
						deploymentCollector.AppendMessage(ref.Name, "Log", sc.Text())
					}
					if err := sc.Err(); err != nil {
						logger.Warn("error streaming logs", zap.Error(err))
						return err
					}
					return nil
				}()
				if err != nil {
					toplevelErr = err
					break resourcesLoop
				}
			}
		}

		if toplevelErr == nil {
			lw.last[d.ID] = now
		}
	}

	if err := lw.logsExporter.Logs(ctx, collector.LogRecords()); err != nil {
		logger.Warn("error exporting logs", zap.Error(err))
	}
}
