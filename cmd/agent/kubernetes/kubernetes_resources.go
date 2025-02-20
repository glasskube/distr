package main

import (
	"bytes"
	"context"
	"errors"
	"io"

	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

func DecodeResourceYaml(data []byte) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(data), 4096)
	var result []*unstructured.Unstructured
	for {
		var object unstructured.Unstructured
		if err := decoder.Decode(&object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		} else if len(object.Object) > 0 {
			result = append(result, &object)
		}
	}
	return result, nil
}

func ApplyResources(ctx context.Context, namespace string, objects []*unstructured.Unstructured) error {
	for _, obj := range objects {
		gvk := obj.GroupVersionKind()
		if mapping, err := k8sRestMapper.RESTMapping(gvk.GroupKind(), gvk.Version); err != nil {
			return err
		} else {
			var resource dynamic.ResourceInterface
			if mapping.Scope.Name() != meta.RESTScopeNameRoot {
				resource = k8sDynamicClient.Resource(mapping.Resource).Namespace(namespace)
			} else {
				resource = k8sDynamicClient.Resource(mapping.Resource)
			}
			if _, err := resource.Get(ctx, obj.GetName(), v1.GetOptions{}); k8serrors.IsNotFound(err) {
				logger.Debug("creating resource",
					zap.String("resourceNamespace", namespace), zap.String("resourceName", obj.GetName()))
				if _, err := resource.Create(ctx, obj, v1.CreateOptions{}); err != nil {
					return err
				}
			} else if err == nil {
				logger.Debug("updating resource",
					zap.String("resourceNamespace", namespace), zap.String("resourceName", obj.GetName()))
				if _, err := resource.Update(ctx, obj, v1.UpdateOptions{}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}
