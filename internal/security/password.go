package security

import (
	"bytes"
	"crypto/rand"
	"errors"
	mathrand "math/rand"

	"github.com/glasskube/cloud/internal/types"
	"golang.org/x/crypto/argon2"
)

const (
	time    = 1
	memory  = 64 * 1024
	threads = 4
	keyLen  = 32
)

var ErrInvalidPassword = errors.New("invalid password")

func HashPassword(userAccount *types.UserAccount) error {
	if salt, err := generateSalt(); err != nil {
		return err
	} else {
		userAccount.PasswordSalt = salt
		userAccount.PasswordHash = generateHash(userAccount.Password, salt)
		userAccount.Password = ""
		return nil
	}
}

func HashAccessKey(accessKey string) ([]byte, []byte, error) {
	if salt, err := generateSalt(); err != nil {
		return nil, nil, err
	} else {
		hash := generateHash(accessKey, salt)
		return salt, hash, nil
	}
}

func VerifyPassword(userAccount types.UserAccount, password string) error {
	if !bytes.Equal(userAccount.PasswordHash, generateHash(password, userAccount.PasswordSalt)) {
		return ErrInvalidPassword
	} else {
		return nil
	}
}

func generateHash(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}

func GenerateAccessKey() string {
	// TODO better
	return randSeq(32)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(b)
}
