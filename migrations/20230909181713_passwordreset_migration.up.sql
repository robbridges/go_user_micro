ALTER TABLE users
    ADD COLUMN password_reset_token text NOT NULL DEFAULT '',
    ADD COLUMN password_reset_expires TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:00';