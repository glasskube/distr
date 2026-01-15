package util_test

import (
	"testing"

	"github.com/distr-sh/distr/internal/util"
	. "github.com/onsi/gomega"
)

func TestMergeIntoRecursive(t *testing.T) {
	g := NewWithT(t)
	a := map[string]any{
		"foo":    "bar",
		"number": 1,
		"only_a": map[string]any{"a": "a"},
		"both":   map[string]any{"aa": "aa"},
	}
	b := map[string]any{
		"foo":    "hello",
		"only_b": map[string]any{"b": "b"},
		"both":   map[string]any{"slice": []any{1, 2, 3}},
	}
	g.Expect(util.MergeIntoRecursive(a, b)).NotTo(HaveOccurred())
	g.Expect(a).To(Equal(map[string]any{
		"foo":    "hello",
		"number": 1,
		"only_a": map[string]any{"a": "a"},
		"only_b": map[string]any{"b": "b"},
		"both":   map[string]any{"aa": "aa", "slice": []any{1, 2, 3}},
	}))
}

func TestMergeIntoRecursive_Error(t *testing.T) {
	g := NewWithT(t)
	a := map[string]any{
		"foo": "bar",
	}
	b := map[string]any{
		"foo": map[string]any{"b": "b"},
	}
	g.Expect(util.MergeIntoRecursive(a, b)).To(HaveOccurred())
}
