package main

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

func GetPodLogsWIP(ctx context.Context, obj runtime.Object) error {
	respMap, err := polymorphichelpers.AllPodLogsForObjectFn(
		k8sConfigFlags,
		obj,
		&corev1.PodLogOptions{
			SinceTime:  &v1.Time{Time: time.Now()},
			Timestamps: true,
		},
		10*time.Second,
		true,
	)
	if err != nil {
		return err
	}
	for _, resp := range respMap {
		_, err := resp.DoRaw(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
