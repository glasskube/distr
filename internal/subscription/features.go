package subscription

import (
	"slices"

	"github.com/glasskube/distr/internal/types"
)

var ProFeatures = []types.Feature{
	types.FeatureLicensing,
}

func AddFeatures(existing []types.Feature, features ...types.Feature) []types.Feature {
	result := slices.Clone(existing)
	for _, f := range features {
		if !slices.Contains(result, f) {
			result = append(result, f)
		}
	}
	return result
}

func RemoveFeatures(existing []types.Feature, features ...types.Feature) []types.Feature {
	result := make([]types.Feature, 0, len(existing))
	for _, f := range existing {
		if !slices.Contains(features, f) {
			result = append(result, f)
		}
	}
	return result
}
