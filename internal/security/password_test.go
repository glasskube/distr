package security_test

import (
	"testing"

	"github.com/distr-sh/distr/internal/security"
	"github.com/distr-sh/distr/internal/types"
	. "github.com/onsi/gomega"
)

func TestHashPassword(t *testing.T) {
	g := NewWithT(t)
	u1 := types.UserAccount{Password: "12345678"}
	u2 := types.UserAccount{Password: "12345678"}
	g.Expect(security.HashPassword(&u1)).NotTo(HaveOccurred())
	g.Expect(u1.Password).To(BeEmpty())
	g.Expect(u1.PasswordSalt).NotTo(BeEmpty())
	g.Expect(security.HashPassword(&u2)).NotTo(HaveOccurred())
	g.Expect(u2.Password).To(BeEmpty())
	g.Expect(u2.PasswordSalt).NotTo(BeEmpty())
	g.Expect(u1.PasswordSalt).NotTo(Equal(u2.PasswordSalt))
	g.Expect(u1.PasswordHash).NotTo(Equal(u2.PasswordHash))
}

func TestVerifyPassword(t *testing.T) {
	g := NewWithT(t)
	pw := "12345678"
	u := types.UserAccount{Password: pw}
	g.Expect(security.HashPassword(&u)).NotTo(HaveOccurred())
	g.Expect(security.VerifyPassword(u, pw)).NotTo(HaveOccurred())
	g.Expect(security.VerifyPassword(u, "wrong")).To(MatchError(security.ErrInvalidPassword))
}
