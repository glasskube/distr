package db

import (
	"context"
	"fmt"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/dbcrypto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EncryptedColumn is one column whose value moved from a plaintext column into an encrypted column
// in migration 130. Rows written before that migration still hold their value in the plaintext
// column until EncryptPlaintextRows has moved it over.
type EncryptedColumn struct {
	Table  string
	Column string
	// Target is the column the encrypted value is written to.
	Target string
	// Binary is true when the plaintext column is a BYTEA rather than a TEXT.
	Binary bool
	// BatchSize is how many rows are read into memory at once. It is small for the tables whose rows
	// can be large.
	BatchSize int
}

// EncryptedColumns is every column the encryption migration covers, in the order it processes them.
var EncryptedColumns = []EncryptedColumn{
	{Table: "Secret", Column: "value", Target: "value_enc"},
	{Table: "CustomOIDCConfiguration", Column: "client_secret", Target: "client_secret_enc"},
	{Table: "CustomEmailConfiguration", Column: "smtp_username", Target: "smtp_username_enc"},
	{Table: "CustomEmailConfiguration", Column: "smtp_password", Target: "smtp_password_enc"},
	{Table: "Artifact", Column: "upstream_username", Target: "upstream_username_enc"},
	{Table: "Artifact", Column: "upstream_password", Target: "upstream_password_enc"},
	{Table: "UserAccount", Column: "mfa_secret", Target: "mfa_secret_enc"},
	{Table: "Organization", Column: "stripe_webhook_secret", Target: "stripe_webhook_secret_enc"},
	{Table: "ApplicationEntitlement", Column: "registry_username", Target: "registry_username_enc"},
	{Table: "ApplicationEntitlement", Column: "registry_password", Target: "registry_password_enc"},
	{Table: "SupportBundle", Column: "bundle_secret", Target: "bundle_secret_enc"},
	{
		Table: "DeploymentRevision", Column: "values_yaml", Target: "values_yaml_enc",
		Binary: true, BatchSize: 200,
	},
	{
		Table: "DeploymentRevision", Column: "env_file_data", Target: "env_file_data_enc",
		Binary: true, BatchSize: 200,
	},
	{Table: "SupportBundleResource", Column: "content", Target: "content_enc", BatchSize: 50},
}

func (c EncryptedColumn) String() string { return c.Table + "." + c.Column }

const defaultEncryptionBatchSize = 500

func (c EncryptedColumn) batchSize() int {
	if c.BatchSize > 0 {
		return c.BatchSize
	}
	return defaultEncryptionBatchSize
}

// plaintextExpr reads the plaintext column as bytes, so that a TEXT and a BYTEA column can be
// encrypted by the same code and produce exactly what the application writes.
func (c EncryptedColumn) plaintextExpr() string {
	if c.Binary {
		return c.Column
	}
	return fmt.Sprintf("convert_to(%s, 'UTF8')", c.Column)
}

// keyPrefixExpr reads the format version and key id a value was sealed with. It has to stay
// identical to the expression migration 130 indexes, and uses substring rather than get_byte
// because that reads a TOAST slice instead of the whole value.
func (c EncryptedColumn) keyPrefixExpr() string {
	return fmt.Sprintf("substring(%s FROM 1 FOR 2)", c.Target)
}

// activeKeyPrefix is the literal that keyPrefixExpr yields for a value sealed with the active key.
func activeKeyPrefix() string {
	return fmt.Sprintf(`'\x%02x%02x'::BYTEA`, dbcrypto.FormatVersion, dbcrypto.Keys().ActiveKeyID())
}

func (c EncryptedColumn) staleKeyExpr() string {
	return fmt.Sprintf("%s <> %s", c.keyPrefixExpr(), activeKeyPrefix())
}

func queryBool(ctx context.Context, query string) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return false, err
	}
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
}

// HasPlaintextRows reports whether any row still holds its value in the plaintext column.
func HasPlaintextRows(ctx context.Context, c EncryptedColumn) (bool, error) {
	found, err := queryBool(ctx,
		fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s IS NOT NULL)", c.Table, c.Column))
	if err != nil {
		return false, fmt.Errorf("could not check %v for plaintext rows: %w", c, err)
	}
	return found, nil
}

// HasStaleKeyRows reports whether any row is still encrypted with a key that is no longer active.
// min and max are answered from the index of migration 130, while a search for a mismatch cannot
// be: no index serves <>, so every plan for it reads the value of every row.
func HasStaleKeyRows(ctx context.Context, c EncryptedColumn) (bool, error) {
	found, err := queryBool(ctx, fmt.Sprintf(
		`SELECT coalesce(min(%[1]s) <> %[2]s OR max(%[1]s) <> %[2]s, false)
		FROM %[3]s WHERE %[4]s IS NOT NULL`,
		c.keyPrefixExpr(), activeKeyPrefix(), c.Table, c.Target))
	if err != nil {
		return false, fmt.Errorf("could not check %v for rows of a retired key: %w", c, err)
	}
	return found, nil
}

type encryptedRow struct {
	ID    uuid.UUID `db:"id"`
	Value []byte    `db:"value"`
}

// EncryptPlaintextRows encrypts every row of one column that is still stored in plaintext and
// returns how many rows it rewrote.
func EncryptPlaintextRows(ctx context.Context, c EncryptedColumn) (int64, error) {
	return rewrite(ctx, c,
		fmt.Sprintf("%s AS value FROM %s WHERE %s IS NOT NULL", c.plaintextExpr(), c.Table, c.Column),
		fmt.Sprintf("%s IS NOT NULL", c.Column),
		dbcrypto.Encrypt,
	)
}

// ReencryptStaleKeyRows re-encrypts every row of one column that was encrypted with a key that is no
// longer active, which is what lets a retired key be removed from the keyring.
func ReencryptStaleKeyRows(ctx context.Context, c EncryptedColumn) (int64, error) {
	return rewrite(ctx, c,
		fmt.Sprintf("%s AS value FROM %s WHERE %s IS NOT NULL AND %s",
			c.Target, c.Table, c.Target, c.staleKeyExpr()),
		c.staleKeyExpr(),
		func(value []byte) ([]byte, error) {
			plaintext, err := dbcrypto.Decrypt(value)
			if err != nil {
				return nil, err
			}
			return dbcrypto.Encrypt(plaintext)
		},
	)
}

// rewrite reads the rows selected by from in batches, re-encrypts each one, and writes the result to
// the target column. Each batch is a statement of its own, so the work can be interrupted and
// resumed, and it can run while the server is serving traffic: guard repeats the selection criteria
// in the update, which makes it a no-op for a row someone else has rewritten since it was read.
func rewrite(
	ctx context.Context,
	c EncryptedColumn,
	from, guard string,
	encrypt func([]byte) ([]byte, error),
) (int64, error) {
	db := internalctx.GetDb(ctx)
	var total int64
	for {
		rows, err := db.Query(ctx, fmt.Sprintf("SELECT id, %s LIMIT %d", from, c.batchSize()))
		if err != nil {
			return total, fmt.Errorf("could not query rows of %v: %w", c, err)
		}
		batch, err := pgx.CollectRows(rows, pgx.RowToStructByName[encryptedRow])
		if err != nil {
			return total, fmt.Errorf("could not collect rows of %v: %w", c, err)
		}
		if len(batch) == 0 {
			return total, nil
		}

		ids := make([]uuid.UUID, len(batch))
		encrypted := make([][]byte, len(batch))
		for i, row := range batch {
			ids[i] = row.ID
			if encrypted[i], err = encrypt(row.Value); err != nil {
				return total, fmt.Errorf("could not encrypt %v of row %v: %w", c, row.ID, err)
			}
		}

		tag, err := db.Exec(ctx,
			fmt.Sprintf(
				`UPDATE %[1]s AS t SET %[2]s = NULL, %[3]s = v.encrypted
				FROM (SELECT unnest(@ids::UUID[]) AS id, unnest(@encrypted::BYTEA[]) AS encrypted) v
				WHERE t.id = v.id AND %[4]s`,
				c.Table, c.Column, c.Target, guard),
			pgx.NamedArgs{"ids": ids, "encrypted": encrypted},
		)
		if err != nil {
			return total, fmt.Errorf("could not encrypt rows of %v: %w", c, err)
		}
		total += tag.RowsAffected()

		if len(batch) < c.batchSize() {
			return total, nil
		}
	}
}
