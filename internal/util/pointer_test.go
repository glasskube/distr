package util_test

import (
	"testing"

	"github.com/distr-sh/distr/internal/util"
	. "github.com/onsi/gomega"
)

func TestPtrEq(t *testing.T) {
	g := NewWithT(t)
	g.Expect(util.PtrEq[any](nil, nil)).To(BeTrue())
	g.Expect(util.PtrEq(util.PtrTo("a"), nil)).To(BeFalse())
	g.Expect(util.PtrEq(nil, util.PtrTo("b"))).To(BeFalse())
	g.Expect(util.PtrEq(util.PtrTo("a"), util.PtrTo("b"))).To(BeFalse())
	g.Expect(util.PtrEq(util.PtrTo("a"), util.PtrTo("a"))).To(BeTrue())
}
