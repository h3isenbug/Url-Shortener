CREATE TABLE IF NOT EXISTS accounts
(
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(64)        NOT NULL
);