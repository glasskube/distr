CREATE TABLE AccessToken (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP DEFAULT current_timestamp,
  expires_at TIMESTAMP,
  last_used_at TIMESTAMP,
  label TEXT,
  key BYTEA UNIQUE NOT NULL,
  user_account_id UUID NOT NULL
    REFERENCES UserAccount (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS AccessToken_key ON AccessToken (key);
CREATE INDEX IF NOT EXISTS fk_AccessToken_user_account_id ON AccessToken (user_account_id);
