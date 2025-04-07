CREATE TYPE TUTORIAL AS ENUM ('branding', 'agents', 'registry');

CREATE TABLE UserAccount_Tutorial (
  useraccount_id UUID NOT NULL REFERENCES UserAccount(id) ON DELETE CASCADE,
  tutorial TUTORIAL NOT NULL,
  data JSONB,
  steps JSONB,
  /* e.g.
steps = {
    "branding": {
      "title": {
        "value": "...",
        "created_at": "..."
      },
      "description": {
        "value": "...",
        "created_at": "..."
      },
    },
    "invite": {
      "email": {
        "value": "...",
        "created_at": "..."
      },
      "customerLogin": {
        "created_at": "..."
      }
    }
}
   */
  created_at TIMESTAMP DEFAULT current_timestamp,
  PRIMARY KEY (useraccount_id, tutorial)
);

CREATE INDEX IF NOT EXISTS UserAccount_Tutorial_useraccount_id ON UserAccount_Tutorial(useraccount_id);
