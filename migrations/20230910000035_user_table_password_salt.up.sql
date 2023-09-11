ALTER TABLE users
    ADD COLUMN password_reset_salt text NOT NULL DEFAULT '';