-- Создание таблицы команд
CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY
);

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT unique_user_id UNIQUE (user_id)
);

-- Создание индекса для быстрого поиска по команде
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Создание таблицы Pull Requests
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP,
    CONSTRAINT unique_pr_id UNIQUE (pull_request_id)
);

-- Создание индексов для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_assigned_reviewers ON pull_requests USING GIN(assigned_reviewers);