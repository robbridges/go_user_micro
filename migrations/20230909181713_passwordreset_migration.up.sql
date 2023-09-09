ALTER TABLE users
    ADD COLUMN password_reset_token text,
    ADD COLUMN password_reset_expiration TIMESTAMP;