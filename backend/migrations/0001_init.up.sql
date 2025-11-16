CREATE TABLE teams
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    deleted_at TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX teams_name_uq_alive
    ON teams (name)
    WHERE deleted_at IS NULL;

CREATE TABLE users
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    VARCHAR(50)  NOT NULL,  -- Внешний идентификатор пользователя.
    name       VARCHAR(150) NOT NULL,
    team_id    BIGINT REFERENCES teams (id) ON DELETE RESTRICT,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ  NULL
);

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pr_status') THEN
            CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
        END IF;
    END
$$;

CREATE TABLE pull_requests
(
    id         BIGSERIAL PRIMARY KEY,
    pr_id      VARCHAR(50)  NOT NULL,  -- Внешний идентификатор pull request.
    title      VARCHAR(100) NOT NULL,
    author_id  BIGINT       REFERENCES users (id) ON DELETE SET NULL,
    status     pr_status    NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL
);

CREATE TABLE pr_reviews
(
    pr_id       BIGINT      NOT NULL REFERENCES pull_requests (id) ON DELETE CASCADE,
    reviewer_id BIGINT      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    slot        SMALLINT    NOT NULL CHECK (slot IN (1, 2)),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pr_id, slot),
    UNIQUE (pr_id, reviewer_id)
);

CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS pr_updated_at_tr ON pull_requests;
CREATE TRIGGER pr_updated_at_tr
    BEFORE UPDATE
    ON pull_requests
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
