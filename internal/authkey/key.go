package authkey

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	keyPrefix       = "distr-"
	secretSeparator = "_"
	saltLength      = 16
)

type Key [16]byte

type Secret [32]byte

// Token is the credential a client sends. Key identifies the AccessToken row and is stored in
// plain text, Secret is what the row's hashes are verified against. Secret is nil for a token
// that was issued before secrets existed, which only authenticates a row that has none.
type Token struct {
	Key    Key
	Secret *Secret
}

var ErrInvalidAccessKey = errors.New("invalid access key")

func Parse(encoded string) (Token, error) {
	if !strings.HasPrefix(encoded, keyPrefix) {
		return Token{}, ErrInvalidAccessKey
	}

	keyPart, secretPart, hasSecret := strings.Cut(strings.TrimPrefix(encoded, keyPrefix), secretSeparator)
	if strings.Contains(secretPart, secretSeparator) {
		return Token{}, ErrInvalidAccessKey
	}

	key, err := parseKey(keyPart)
	if err != nil {
		return Token{}, err
	} else if !hasSecret {
		return Token{Key: key}, nil
	}

	secret, err := parseSecret(secretPart)
	if err != nil {
		return Token{}, err
	}
	return Token{Key: key, Secret: &secret}, nil
}

func NewToken() (Token, error) {
	key, err := NewKey()
	if err != nil {
		return Token{}, err
	}
	secret, err := NewSecret()
	if err != nil {
		return Token{}, err
	}
	return Token{Key: key, Secret: &secret}, nil
}

func (token Token) String() string { return token.Key.String() }

func (token Token) Serialize() string {
	if token.Secret == nil {
		return token.Key.Serialize()
	}
	return token.Key.Serialize() + secretSeparator + hex.EncodeToString(token.Secret[:])
}

func NewKey() (key Key, err error) {
	_, err = rand.Read(key[:])
	return key, err
}

func parseKey(encoded string) (Key, error) {
	if decoded, err := hex.DecodeString(encoded); err != nil {
		return Key{}, fmt.Errorf("%w: %w", ErrInvalidAccessKey, err)
	} else if len(decoded) != len(Key{}) {
		return Key{}, ErrInvalidAccessKey
	} else {
		return Key(decoded), nil
	}
}

func (key Key) String() string {
	return keyPrefix + hex.EncodeToString(key[:3]) + "___REDACTED___"
}

func (key Key) Serialize() string { return keyPrefix + hex.EncodeToString(key[:]) }

func (key *Key) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		if len(v) == len(Key{}) {
			*key = Key(slices.Clone(v))
			return nil
		}
	}
	return errors.New("cannot scan into Key")
}

func NewSecret() (secret Secret, err error) {
	_, err = rand.Read(secret[:])
	return secret, err
}

func parseSecret(encoded string) (Secret, error) {
	if decoded, err := hex.DecodeString(encoded); err != nil {
		return Secret{}, fmt.Errorf("%w: %w", ErrInvalidAccessKey, err)
	} else if len(decoded) != len(Secret{}) {
		return Secret{}, ErrInvalidAccessKey
	} else {
		return Secret(decoded), nil
	}
}

func (secret Secret) String() string { return "___REDACTED___" }

// Hash derives the value stored in the database. Unlike a password, a secret is 256 bits of
// CSPRNG output, so there is no dictionary to search and no need for a costly key derivation
// function: one HMAC-SHA256 keeps verification affordable on every single request.
func (secret Secret) Hash(salt []byte) []byte {
	mac := hmac.New(sha256.New, salt)
	mac.Write(secret[:])
	return mac.Sum(nil)
}

func NewSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	return salt, err
}

func VerifySecret(salt, hash []byte, secret Secret) bool {
	return subtle.ConstantTimeCompare(hash, secret.Hash(salt)) == 1
}
