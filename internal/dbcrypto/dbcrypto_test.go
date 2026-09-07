package dbcrypto

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/glasskube/pkg/crypto"
	"github.com/jackc/pgx/v5/pgtype"
	. "github.com/onsi/gomega"
)

func testKey(fill byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{fill}, crypto.MinKeyLen))
}

// useKeyring points the process wide keyring at a test keyring, so the [String] and [Bytes] scan
// paths can run without an initialized environment.
func useKeyring(t *testing.T) {
	t.Helper()
	keyring, err := crypto.ParseKeyring(testKey(1))
	NewWithT(t).Expect(err).NotTo(HaveOccurred())
	previous := keys
	keys = keyring
	t.Cleanup(func() { keys = previous })
}

func TestEncryptDecrypt(t *testing.T) {
	t.Run("round trips a value", func(t *testing.T) {
		g := NewWithT(t)
		useKeyring(t)
		sealed, err := Encrypt([]byte("hunter2"))
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(sealed)).NotTo(ContainSubstring("hunter2"))
		g.Expect(Decrypt(sealed)).To(Equal([]byte("hunter2")))
	})

	t.Run("compresses a large value, which a column can no longer do for itself", func(t *testing.T) {
		g := NewWithT(t)
		useKeyring(t)
		plaintext := []byte(strings.Repeat("deployment log line\n", 1000))
		sealed, err := Encrypt(plaintext)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(sealed)).To(BeNumerically("<", len(plaintext)))
		g.Expect(Decrypt(sealed)).To(Equal(plaintext))
	})

	t.Run("returns a plaintext marked value unchanged", func(t *testing.T) {
		g := NewWithT(t)
		useKeyring(t)
		g.Expect(Decrypt(append([]byte{plaintextMarker}, "hunter2"...))).To(Equal([]byte("hunter2")))
	})

	t.Run("rejects an empty value", func(t *testing.T) {
		g := NewWithT(t)
		useKeyring(t)
		_, err := Decrypt(nil)
		g.Expect(err).To(HaveOccurred())
	})
}

// TestScanPlan pins down that pgx routes an encrypted column into the decrypting Scan method rather
// than assigning the raw ciphertext to the underlying string or byte slice.
func TestScanPlan(t *testing.T) {
	useKeyring(t)
	sealed, err := Encrypt([]byte("hunter2"))
	NewWithT(t).Expect(err).NotTo(HaveOccurred())
	m := pgtype.NewMap()

	t.Run("String", func(t *testing.T) {
		g := NewWithT(t)
		var dst String
		plan := m.PlanScan(pgtype.ByteaOID, pgtype.BinaryFormatCode, &dst)
		g.Expect(plan).NotTo(BeNil())
		g.Expect(plan.Scan(sealed, &dst)).To(Succeed())
		g.Expect(dst).To(Equal(String("hunter2")))
	})

	t.Run("nullable String", func(t *testing.T) {
		g := NewWithT(t)
		var dst *String
		plan := m.PlanScan(pgtype.ByteaOID, pgtype.BinaryFormatCode, &dst)
		g.Expect(plan).NotTo(BeNil())
		g.Expect(plan.Scan(sealed, &dst)).To(Succeed())
		g.Expect(dst).To(HaveValue(Equal(String("hunter2"))))

		g.Expect(plan.Scan(nil, &dst)).To(Succeed())
		g.Expect(dst).To(BeNil())
	})

	t.Run("Bytes", func(t *testing.T) {
		g := NewWithT(t)
		var dst Bytes
		plan := m.PlanScan(pgtype.ByteaOID, pgtype.BinaryFormatCode, &dst)
		g.Expect(plan).NotTo(BeNil())
		g.Expect(plan.Scan(sealed, &dst)).To(Succeed())
		g.Expect(dst).To(Equal(Bytes("hunter2")))

		g.Expect(plan.Scan(nil, &dst)).To(Succeed())
		g.Expect(dst).To(BeNil())
	})
}
