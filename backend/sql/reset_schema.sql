-- 清空并重建库表（仅结构，无业务数据）。执行顺序：先本文件，再 seed_default.sql
-- 注意：会删除全部用户、板块、帖子、评论、投票

DROP TABLE IF EXISTS "comment_vote";
DROP TABLE IF EXISTS "comment_report";
DROP TABLE IF EXISTS "comment";
DROP TABLE IF EXISTS "post_tag";
DROP TABLE IF EXISTS "tag";
DROP TABLE IF EXISTS "post_vote";
DROP TABLE IF EXISTS "post_favorite";
DROP TABLE IF EXISTS "post_report";
DROP TABLE IF EXISTS "post_appeal";
DROP TABLE IF EXISTS "moderation_log";
DROP TABLE IF EXISTS "post";
DROP TABLE IF EXISTS "board_favorite";
DROP TABLE IF EXISTS "board_moderator";
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
    visibility VARCHAR(16) NOT NULL DEFAULT 'public',
    is_system_sink BOOLEAN NOT NULL DEFAULT FALSE,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(slug, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(description, '')), 'C')
    ) STORED,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_slug_unique UNIQUE (slug),
    CONSTRAINT board_visibility_check CHECK (visibility IN ('public', 'private')),
    CONSTRAINT board_created_by_fk FOREIGN KEY (created_by) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX idx_board_created_by ON "board" (created_by);
CREATE INDEX idx_board_visibility ON "board" (visibility);

-- 板块版主（创建者写入首行，后续可扩展任命）
CREATE TABLE "board_moderator" (
    user_id BIGINT NOT NULL,
    board_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'moderator',
    appointed_by BIGINT,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_moderator_pk PRIMARY KEY (user_id, board_id),
    CONSTRAINT board_moderator_role_check CHECK (role IN ('owner', 'moderator')),
    CONSTRAINT board_moderator_user_fk FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT board_moderator_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE,
    CONSTRAINT board_moderator_appointed_by_fk FOREIGN KEY (appointed_by) REFERENCES "user" (user_id) ON DELETE SET NULL
);
CREATE INDEX idx_board_moderator_board_id ON "board_moderator" (board_id);
CREATE INDEX idx_board_moderator_board_role ON "board_moderator" (board_id, role);
CREATE INDEX idx_board_search_vector ON "board" USING GIN (search_vector);

-- 用户收藏板块（订阅/星标）
CREATE TABLE "board_favorite" (
    user_id BIGINT NOT NULL,
    board_id BIGINT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_favorite_pk PRIMARY KEY (user_id, board_id),
    CONSTRAINT board_favorite_user_fk FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT board_favorite_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE
);
CREATE INDEX idx_board_favorite_board_id ON "board_favorite" (board_id);
CREATE INDEX idx_board_favorite_user_time ON "board_favorite" (user_id, create_time DESC);

CREATE TABLE "post" (
    id BIGSERIAL PRIMARY KEY,
    board_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT,
    deleted_at TIMESTAMPTZ,
    sealed_at TIMESTAMPTZ,
    sealed_by_user_id BIGINT,
    seal_kind VARCHAR(16),
    comments_locked_at TIMESTAMPTZ,
    comments_locked_by_user_id BIGINT,
    pinned_at TIMESTAMPTZ,
    pinned_by_user_id BIGINT,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(content, '')), 'B')
    ) STORED,
    score BIGINT NOT NULL DEFAULT 0,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_board_id_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE RESTRICT,
    CONSTRAINT post_author_id_foreign FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_sealed_by_fk FOREIGN KEY (sealed_by_user_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_comments_locked_by_fk FOREIGN KEY (comments_locked_by_user_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_pinned_by_fk FOREIGN KEY (pinned_by_user_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_seal_kind_check CHECK (seal_kind IS NULL OR seal_kind IN ('moderator', 'site'))
);
CREATE INDEX idx_post_board_id ON "post" (board_id);
CREATE INDEX idx_post_board_create_time ON "post" (board_id, create_time DESC);
CREATE INDEX idx_post_author_id ON "post" (author_id);
CREATE INDEX idx_post_search_vector ON "post" USING GIN (search_vector) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_board_score ON "post" (board_id, score DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_board_pinned_at ON "post" (board_id, pinned_at DESC) WHERE deleted_at IS NULL;

-- 帖子举报：由用户上报，版主/站主管理处理状态。
CREATE TABLE "post_report" (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    board_id BIGINT NOT NULL,
    reporter_id BIGINT NOT NULL,
    reason VARCHAR(120) NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    handler_id BIGINT,
    handler_note TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_report_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT post_report_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE,
    CONSTRAINT post_report_reporter_fk FOREIGN KEY (reporter_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT post_report_handler_fk FOREIGN KEY (handler_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_report_status_check CHECK (status IN ('open', 'in_review', 'resolved', 'rejected'))
);
CREATE INDEX idx_post_report_board_status_time ON "post_report" (board_id, status, create_time DESC);
CREATE INDEX idx_post_report_post_time ON "post_report" (post_id, create_time DESC);
CREATE INDEX idx_post_report_reporter_time ON "post_report" (reporter_id, create_time DESC);

-- 封帖申诉：作者可对被封帖子提交修改申请，版主处理并回复。
CREATE TABLE "post_appeal" (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    board_id BIGINT NOT NULL,
    author_id BIGINT NOT NULL,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    requested_title VARCHAR(300) NOT NULL,
    requested_content TEXT NOT NULL DEFAULT '',
    user_reply TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    moderator_id BIGINT,
    moderator_reply TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_appeal_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT post_appeal_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE,
    CONSTRAINT post_appeal_author_fk FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT post_appeal_moderator_fk FOREIGN KEY (moderator_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT post_appeal_status_check CHECK (status IN ('open', 'in_review', 'approved', 'rejected'))
);
CREATE INDEX idx_post_appeal_board_status_time ON "post_appeal" (board_id, status, create_time DESC);
CREATE INDEX idx_post_appeal_author_post_time ON "post_appeal" (author_id, post_id, create_time DESC);

-- 治理日志：记录版主/站管的治理行为，用于审计追踪。
CREATE TABLE "moderation_log" (
    id BIGSERIAL PRIMARY KEY,
    board_id BIGINT NOT NULL,
    operator_id BIGINT NOT NULL,
    action VARCHAR(48) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT moderation_log_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE,
    CONSTRAINT moderation_log_operator_fk FOREIGN KEY (operator_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT moderation_log_action_check CHECK (action IN (
        'seal_post', 'unseal_post', 'delete_post', 'restore_post',
        'handle_post_appeal',
        'lock_post_comments', 'unlock_post_comments',
        'pin_post', 'unpin_post',
        'update_post_report_status', 'update_comment_report_status',
        'upsert_board_moderator', 'update_board_moderator_role', 'remove_board_moderator'
    )),
    CONSTRAINT moderation_log_target_type_check CHECK (target_type IN ('post', 'post_report', 'comment_report', 'board_moderator'))
);
CREATE INDEX idx_moderation_log_board_time ON "moderation_log" (board_id, create_time DESC);
CREATE INDEX idx_moderation_log_board_action_time ON "moderation_log" (board_id, action, create_time DESC);
CREATE INDEX idx_moderation_log_operator_time ON "moderation_log" (operator_id, create_time DESC);

-- 用户收藏帖子（星标）
CREATE TABLE "post_favorite" (
    user_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_favorite_pk PRIMARY KEY (user_id, post_id),
    CONSTRAINT post_favorite_user_fk FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT post_favorite_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE
);
CREATE INDEX idx_post_favorite_post_id ON "post_favorite" (post_id);
CREATE INDEX idx_post_favorite_user_time ON "post_favorite" (user_id, create_time DESC);

-- 每用户每帖至多一行；value=1 上票，-1 下票；改票 UPDATE，取消 DELETE
CREATE TABLE "post_vote" (
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    value SMALLINT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_vote_pk PRIMARY KEY (post_id, user_id),
    CONSTRAINT post_vote_value_check CHECK (value IN (-1, 1)),
    CONSTRAINT post_vote_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT post_vote_user_fk FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE
);
CREATE INDEX idx_post_vote_user_id ON "post_vote" (user_id);

-- 标签：全站共用
CREATE TABLE "tag" (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    description TEXT,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT tag_slug_unique UNIQUE (slug)
);
CREATE INDEX idx_tag_name ON "tag" (name);

-- 帖子-标签关联（多对多）
CREATE TABLE "post_tag" (
    post_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_tag_pk PRIMARY KEY (post_id, tag_id),
    CONSTRAINT post_tag_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT post_tag_tag_fk FOREIGN KEY (tag_id) REFERENCES "tag" (id) ON DELETE RESTRICT
);
CREATE INDEX idx_post_tag_tag_id ON "post_tag" (tag_id);

-- 评论：挂在帖子下；parent_id 为空表示顶层，非空表示回复某条评论（楼中楼）
CREATE TABLE "comment" (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    author_id BIGINT,
    parent_id BIGINT,
    content TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    score BIGINT NOT NULL DEFAULT 0,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT comment_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT comment_author_fk FOREIGN KEY (author_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT comment_parent_fk FOREIGN KEY (parent_id) REFERENCES "comment" (id) ON DELETE CASCADE
);
CREATE INDEX idx_comment_post_id ON "comment" (post_id);
CREATE INDEX idx_comment_post_create_time ON "comment" (post_id, create_time);
CREATE INDEX idx_comment_parent_id ON "comment" (parent_id);

-- 评论举报：与帖子举报状态机一致，按板块治理。
CREATE TABLE "comment_report" (
    id BIGSERIAL PRIMARY KEY,
    comment_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    board_id BIGINT NOT NULL,
    reporter_id BIGINT NOT NULL,
    reason VARCHAR(120) NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    handler_id BIGINT,
    handler_note TEXT NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT comment_report_comment_fk FOREIGN KEY (comment_id) REFERENCES "comment" (id) ON DELETE CASCADE,
    CONSTRAINT comment_report_post_fk FOREIGN KEY (post_id) REFERENCES "post" (id) ON DELETE CASCADE,
    CONSTRAINT comment_report_board_fk FOREIGN KEY (board_id) REFERENCES "board" (id) ON DELETE CASCADE,
    CONSTRAINT comment_report_reporter_fk FOREIGN KEY (reporter_id) REFERENCES "user" (user_id) ON DELETE CASCADE,
    CONSTRAINT comment_report_handler_fk FOREIGN KEY (handler_id) REFERENCES "user" (user_id) ON DELETE SET NULL,
    CONSTRAINT comment_report_status_check CHECK (status IN ('open', 'in_review', 'resolved', 'rejected'))
);
CREATE INDEX idx_comment_report_board_status_time ON "comment_report" (board_id, status, create_time DESC);
CREATE INDEX idx_comment_report_comment_time ON "comment_report" (comment_id, create_time DESC);
CREATE INDEX idx_comment_report_reporter_time ON "comment_report" (reporter_id, create_time DESC);

-- 评论投票：语义同 post_vote
CREATE TABLE "comment_vote" (
    comment_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    value SMALLINT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT comment_vote_pk PRIMARY KEY (comment_id, user_id),
    CONSTRAINT comment_vote_value_check CHECK (value IN (-1, 1)),
    CONSTRAINT comment_vote_comment_fk FOREIGN KEY (comment_id) REFERENCES "comment" (id) ON DELETE CASCADE,
    CONSTRAINT comment_vote_user_fk FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE
);
CREATE INDEX idx_comment_vote_user_id ON "comment_vote" (user_id);
