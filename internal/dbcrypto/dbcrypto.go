package dbcrypto

import "fmt"

var keys *Keyring

// Init parses the given DATABASE_ENCRYPTION_KEY value into the keyring of this instance. Every
// command that reads or writes an encrypted column has to call it before it does, so that a
// malformed key aborts startup instead of failing the first query that touches such a column.
func Init(spec string) error {
	keyring, err := ParseKeyring(spec)
	if err != nil {
		return err
	}
	keys = keyring
	return nil
}

// Keys MUST be called after [Init], otherwise it WILL panic.
func Keys() *Keyring {
	if keys == nil {
		panic("detected call to dbcrypto.Keys before calling dbcrypto.Init")
	}
	return keys
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
