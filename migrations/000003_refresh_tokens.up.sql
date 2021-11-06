CREATE SEQUENCE IF NOT EXISTS refresh_token_family MINVALUE 1;

CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id          SERIAL PRIMARY KEY,
    account_id  INTEGER                  NOT NULL REFERENCES accounts (id),
    token       VARCHAR(256)             NOT NULL UNIQUE,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    compromised BOOLEAN                  NOT NULL DEFAULT FALSE,
    disabled    BOOLEAN                  NOT NULL DEFAULT FALSE,
    family      INTEGER                  NOT NULL DEFAULT nextval('refresh_token_family'),
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS refresh_token_family ON refresh_tokens USING btree (family);
