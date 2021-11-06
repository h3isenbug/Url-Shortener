CREATE TABLE IF NOT EXISTS urls
(
    id            SERIAL PRIMARY KEY,
    original_url  VARCHAR(2048)            NOT NULL,
    slug          VARCHAR(40)              NOT NULL,
    total_visits  BIGINT                   NOT NULL DEFAULT 0,
    unique_visits BIGINT                   NOT NULL DEFAULT 0,
    disabled      BOOLEAN                  NOT NULL DEFAULT FALSE,

    account_id    INTEGER                  NOT NULL REFERENCES accounts (id),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
