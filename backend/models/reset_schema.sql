-- 清空并重建库表（仅结构，无业务数据）。执行顺序：先本文件，再 seed_default.sql
-- 注意：会删除全部用户、板块、帖子、评论

DROP TABLE IF EXISTS "comment";
DROP TABLE IF EXISTS "post";
DROP TABLE IF EXISTS "board";
DROP TABLE IF EXISTS "user";

CREATE TABLE "user" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(64) NOT NULL,
    email VARCHAR(64),
    is_site_admin BOOLEAN NOT NULL DEFAULT FALSE,
    gender SMALLINT NOT NULL DEFAULT 0,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_username_unique UNIQUE (username),
    CONSTRAINT user_user_id_unique UNIQUE (user_id)
);

CREATE TABLE "board" (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    created_by BIGINT,
    is_system_sink BOOLEAN NOT NULL DEFAULT FALSE,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(slug, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(description, '')), 'C')
    ) STORED,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_slug_unique UNIQUE (slug),
    CONSTRAINT board_created_by_fk FOREIGN KEY (created_by) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX idx_board_created_by ON "board" (created_by);
CREATE INDEX idx_board_search_vector ON "board" USING GIN (search_vector);

CREATE TABLE "post" (
    id BIGSERIAL PRIMARY KEY,
    board_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT,
    deleted_at TIMESTAMPTZ,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(content, '')), 'B')
    ) STORED,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_board_id_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE RESTRICT,
    CONSTRAINT post_author_id_foreign FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX idx_post_board_id ON "post" (board_id);
CREATE INDEX idx_post_board_create_time ON "post" (board_id, create_time DESC);
CREATE INDEX idx_post_author_id ON "post" (author_id);
CREATE INDEX idx_post_search_vector ON "post" USING GIN (search_vector) WHERE deleted_at IS NULL;

-- 评论：挂在帖子下；parent_id 为空表示顶层，非空表示回复某条评论（楼中楼）
CREATE TABLE "comment" (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    author_id BIGINT,
    parent_id BIGINT,
    content TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT comment_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT comment_author_fk FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT comment_parent_fk FOREIGN KEY (parent_id) REFERENCES "comment" (id) ON DELETE CASCADE
);
CREATE INDEX idx_comment_post_id ON "comment" (post_id);
CREATE INDEX idx_comment_post_create_time ON "comment" (post_id, create_time);
CREATE INDEX idx_comment_parent_id ON "comment" (parent_id);
