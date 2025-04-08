CREATE TYPE TUTORIAL AS ENUM ('branding', 'agents', 'registry');

CREATE TABLE UserAccount_Tutorial (
  useraccount_id UUID NOT NULL REFERENCES UserAccount(id) ON DELETE CASCADE,
  tutorial TUTORIAL NOT NULL,
  events JSONB,
  created_at TIMESTAMP DEFAULT current_timestamp,
  completed_at TIMESTAMP,
  PRIMARY KEY (useraccount_id, tutorial)
);

CREATE INDEX IF NOT EXISTS UserAccount_Tutorial_useraccount_id ON UserAccount_Tutorial(useraccount_id);
