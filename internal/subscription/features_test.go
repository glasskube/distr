package subscription_test

import (
	"testing"

	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	. "github.com/onsi/gomega"
)

func TestAddFeatures(t *testing.T) {
	tests := []struct {
		name     string
		existing []types.Feature
		toAdd    []types.Feature
		want     []types.Feature
	}{
		{
			name:     "add to empty list",
			existing: []types.Feature{},
			toAdd:    []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{types.FeatureLicensing},
		},
		{
			name:     "add to existing list",
			existing: []types.Feature{types.FeatureLicensing},
			toAdd:    []types.Feature{types.FeaturePrePostScripts},
			want:     []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
		},
		{
			name:     "add duplicate feature",
			existing: []types.Feature{types.FeatureLicensing},
			toAdd:    []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{types.FeatureLicensing},
		},
		{
			name:     "add multiple features",
			existing: []types.Feature{types.FeatureLicensing},
			toAdd:    []types.Feature{types.FeaturePrePostScripts, types.FeatureLicensing},
			want:     []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
		},
		{
			name:     "add nothing",
			existing: []types.Feature{types.FeatureLicensing},
			toAdd:    []types.Feature{},
			want:     []types.Feature{types.FeatureLicensing},
		},
		{
			name:     "preserve order when adding",
			existing: []types.Feature{types.FeaturePrePostScripts, types.FeatureLicensing},
			toAdd:    []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{types.FeaturePrePostScripts, types.FeatureLicensing},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := subscription.AddFeatures(tt.existing, tt.toAdd...)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestRemoveFeatures(t *testing.T) {
	tests := []struct {
		name     string
		existing []types.Feature
		toRemove []types.Feature
		want     []types.Feature
	}{
		{
			name:     "remove from list",
			existing: []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
			toRemove: []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{types.FeaturePrePostScripts},
		},
		{
			name:     "remove non-existent feature",
			existing: []types.Feature{types.FeatureLicensing},
			toRemove: []types.Feature{types.FeaturePrePostScripts},
			want:     []types.Feature{types.FeatureLicensing},
		},
		{
			name:     "remove all features",
			existing: []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
			toRemove: []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
			want:     []types.Feature{},
		},
		{
			name:     "remove from empty list",
			existing: []types.Feature{},
			toRemove: []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{},
		},
		{
			name:     "remove nothing",
			existing: []types.Feature{types.FeatureLicensing},
			toRemove: []types.Feature{},
			want:     []types.Feature{types.FeatureLicensing},
		},
		{
			name:     "preserve order when removing",
			existing: []types.Feature{types.FeaturePrePostScripts, types.FeatureLicensing},
			toRemove: []types.Feature{types.FeatureLicensing},
			want:     []types.Feature{types.FeaturePrePostScripts},
		},
		{
			name:     "remove multiple features",
			existing: []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts},
			toRemove: []types.Feature{types.FeaturePrePostScripts, types.FeatureLicensing},
			want:     []types.Feature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := subscription.RemoveFeatures(tt.existing, tt.toRemove...)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestAddFeatures_DoesNotMutateInput(t *testing.T) {
	g := NewWithT(t)
	original := []types.Feature{types.FeatureLicensing}
	originalCopy := make([]types.Feature, len(original))
	copy(originalCopy, original)

	subscription.AddFeatures(original, types.FeaturePrePostScripts)

	g.Expect(original).To(Equal(originalCopy))
}

func TestRemoveFeatures_DoesNotMutateInput(t *testing.T) {
	g := NewWithT(t)
	original := []types.Feature{types.FeatureLicensing, types.FeaturePrePostScripts}
	originalCopy := make([]types.Feature, len(original))
	copy(originalCopy, original)

	subscription.RemoveFeatures(original, types.FeatureLicensing)

	g.Expect(original).To(Equal(originalCopy))
}
