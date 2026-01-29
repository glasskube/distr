ALTER TABLE UserAccount
  ADD COLUMN mfa_secret TEXT,
  ADD COLUMN mfa_enabled BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN mfa_enabled_at TIMESTAMP WITH TIME ZONE,
  ADD CONSTRAINT mfa_secret_not_null_if_enabled
    CHECK (mfa_enabled = false OR mfa_secret IS NOT NULL);
