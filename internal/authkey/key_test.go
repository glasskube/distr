package authkey

import (
	"encoding/hex"
	"testing"

	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewWithT(t)

	token, err := NewToken()
	g.Expect(err).ToNot(HaveOccurred())

	parsed, err := Parse(token.Serialize())
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(parsed.Key).To(Equal(token.Key))
	g.Expect(parsed.Secret).To(HaveValue(Equal(*token.Secret)))

	// Tokens issued before secrets existed carry the key only.
	legacy, err := Parse(token.Key.Serialize())
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(legacy.Key).To(Equal(token.Key))
	g.Expect(legacy.Secret).To(BeNil())

	key := hex.EncodeToString(token.Key[:])
	secret := hex.EncodeToString(token.Secret[:])
	for name, encoded := range map[string]string{
		"no prefix":         key + "_" + secret,
		"key too short":     "distr-" + key[2:] + "_" + secret,
		"key not hex":       "distr-" + "zz" + key[2:] + "_" + secret,
		"secret too short":  "distr-" + key + "_" + secret[2:],
		"secret not hex":    "distr-" + key + "_zz" + secret[2:],
		"empty secret":      "distr-" + key + "_",
		"second separator":  "distr-" + key + "_" + secret + "_" + secret,
		"separator in key":  "distr-" + "_" + key + "_" + secret,
		"nothing but prefix": "distr-",
	} {
		_, err := Parse(encoded)
		g.Expect(err).To(MatchError(ErrInvalidAccessKey), name)
	}
}

func TestVerifySecret(t *testing.T) {
	g := NewWithT(t)

	secret, err := NewSecret()
	g.Expect(err).ToNot(HaveOccurred())
	other, err := NewSecret()
	g.Expect(err).ToNot(HaveOccurred())
	salt, err := NewSalt()
	g.Expect(err).ToNot(HaveOccurred())
	otherSalt, err := NewSalt()
	g.Expect(err).ToNot(HaveOccurred())

	hash := secret.Hash(salt)
	g.Expect(hash).ToNot(Equal(secret[:]))
	g.Expect(VerifySecret(salt, hash, secret)).To(BeTrue())
	g.Expect(VerifySecret(salt, hash, other)).To(BeFalse())
	g.Expect(VerifySecret(otherSalt, hash, secret)).To(BeFalse())
}
