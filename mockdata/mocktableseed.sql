CREATE TABLE IF NOT EXISTS users (
     id SERIAL PRIMARY KEY,
     password_hash text NOT NULL,
     email text UNIQUE NOT NULL,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
     password_reset_token text,
     password_reset_expires TIMESTAMP,
     password_reset_salt text
);

INSERT INTO users (password_hash, email) VALUES ('$2a$10$m2RvoCSnhAMGZggN1SPPsOwlSC8Ne0EX.wi7EHK2/pKKmoOmDQsUe', 'admin@localhost');
