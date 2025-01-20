package authkey

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const keyPrefix = "gkc-"

type Key [16]byte

var ErrInvalidAccessKey = errors.New("invalid access key")

func Parse(encoded string) (Key, error) {
	if !strings.HasPrefix(encoded, keyPrefix) {
		return Key{}, ErrInvalidAccessKey
	} else if decoded, err := hex.DecodeString(strings.TrimPrefix(encoded, keyPrefix)); err != nil {
		return Key{}, fmt.Errorf("%w: %w", ErrInvalidAccessKey, err)
	} else {
		return Key(decoded), nil
	}
}

func NewKey() (key Key, err error) {
	_, err = rand.Read(key[:])
	return
}

func (key Key) Serialize() string { return keyPrefix + hex.EncodeToString(key[:]) }

func (key Key) MarshalJSON() ([]byte, error) { return json.Marshal(key.Serialize()) }
