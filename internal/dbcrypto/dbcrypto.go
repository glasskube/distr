package dbcrypto

import (
	"bytes"
	"fmt"

	"github.com/glasskube/pkg/crypto"
)

// FormatVersion is the first byte of every encrypted value, and the key id is the second. A
// maintenance query relies on this to find the values of a retired key without decrypting them.
const FormatVersion = crypto.FormatVersion

// plaintextMarker prefixes a value that a query read from the plaintext column of a row the
// encryption migration has not moved over yet, which is how a read tells the two apart without a
// second column in the result. [crypto.FormatVersion] 0 is reserved upstream for exactly this, so a
// marked value can never be mistaken for an encrypted one.
const plaintextMarker byte = 0

var keys *crypto.Keyring

// Init parses the given DATABASE_ENCRYPTION_KEY value into the keyring of this instance. Every
// command that reads or writes an encrypted column has to call it before it does, so that a
// malformed key aborts startup instead of failing the first query that touches such a column. The
// key is passed in because internal/types imports this package, and through it every agent binary,
// which has no internal/env.
func Init(spec string) error {
	keyring, err := crypto.ParseKeyring(spec)
	if err != nil {
		return err
	}
	keys = keyring
	return nil
}

// Keys MUST be called after [Init], otherwise it WILL panic.
func Keys() *crypto.Keyring {
	if keys == nil {
		panic("detected call to dbcrypto.Keys before calling dbcrypto.Init")
	}
	return keys
}

// Encrypt seals a value of an encrypted column with the active key. Compression is on because
// Postgres can no longer TOAST-compress a column once it holds ciphertext, and the columns this
// package covers include support bundle resources and deployment revision values, which are large
// enough for that to matter.
func Encrypt(plaintext []byte) ([]byte, error) {
	return Keys().Encrypt(plaintext, crypto.WithCompression())
}

// Decrypt opens a value of an encrypted column, and returns a value that is still stored in a
// plaintext column unchanged.
func Decrypt(value []byte) ([]byte, error) {
	if len(value) > 0 && value[0] == plaintextMarker {
		// pgx may reuse the buffer a row was scanned from, so the result must not alias it.
		return bytes.Clone(value[1:]), nil
	}
	return Keys().Decrypt(value)
}

// TextValue renders the expression that reads an encrypted TEXT column. Until the encryption
// migration has run, a row may still hold its value in the plaintext column, so both are merged into
// one value that a [String] can scan. It carries no column alias, because an output expression is
// also used inside a row constructor, where an alias is a syntax error.
func TextValue(alias, column string) string {
	return fmt.Sprintf(
		`coalesce(%[1]s.%[2]s_enc, '\x00'::bytea || convert_to(%[1]s.%[2]s, 'UTF8'))`,
		alias, column,
	)
}

// BytesValue is [TextValue] for a BYTEA column, scannable into a [Bytes].
func BytesValue(alias, column string) string {
	return fmt.Sprintf(`coalesce(%[1]s.%[2]s_enc, '\x00'::bytea || %[1]s.%[2]s)`, alias, column)
}

// IsSetValue reports whether either representation of an encrypted column holds a value, for the
// booleans an API exposes in place of a secret it must not return.
func IsSetValue(alias, column string) string {
	return fmt.Sprintf(`num_nonnulls(%[1]s.%[2]s, %[1]s.%[2]s_enc) > 0`, alias, column)
}

// TextColumn is [TextValue] named after the column, for a result that is scanned by name.
func TextColumn(alias, column string) string {
	return TextValue(alias, column) + " AS " + column
}

// BytesColumn is [BytesValue] named after the column, for a result that is scanned by name.
func BytesColumn(alias, column string) string {
	return BytesValue(alias, column) + " AS " + column
}
