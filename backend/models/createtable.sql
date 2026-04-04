CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(64) NOT NULL,
    email VARCHAR(64),
    gender SMALLINT NOT NULL DEFAULT 0,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_username_unique UNIQUE (username),
    CONSTRAINT user_user_id_unique UNIQUE (user_id)
);

-- 板块（仿 Subreddit）。删除某板块前，应将其下帖子 board_id 更新到 is_system_sink 归档板，再删原板块行。
CREATE TABLE IF NOT EXISTS "board" (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    created_by BIGINT,
    is_system_sink BOOLEAN NOT NULL DEFAULT FALSE,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_slug_unique UNIQUE (slug),
    CONSTRAINT board_created_by_fk FOREIGN KEY (created_by) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_board_created_by ON "board" (created_by);

-- 系统归档板：板块被删除时，其帖子统一 UPDATE 到此 board.id（slug 固定 _archived，勿删勿改 slug 语义）。
-- general：默认讨论板，新业务在尚未选板前可指向此板，避免新帖落在「已归档」名下。
INSERT INTO "board" (slug, name, description, is_system_sink)
VALUES
    ('_archived', '已归档', '原板块删除后，帖子统一归并到此板块。', TRUE),
    ('general', '综合', '默认讨论板块。', FALSE)
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE IF NOT EXISTS "post" (
    id BIGSERIAL PRIMARY KEY,
    board_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_board_id_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE RESTRICT,
    CONSTRAINT post_author_id_foreign FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_post_board_id ON "post" (board_id);
CREATE INDEX IF NOT EXISTS idx_post_board_create_time ON "post" (board_id, create_time DESC);
CREATE INDEX IF NOT EXISTS idx_post_author_id ON "post" (author_id);
