-- 默认种子数据。请先执行 reset_schema.sql 再执行本文件。
--
-- 登录说明（密码与后端 dao 中 encryptPassword 一致：MD5(secret+username+明文)，secret=hyanhsing0101）：
--   hyan  / changyears45  -> 你提供的哈希 c6f3780db708c2ad91a343b14c4e3b77（仅对用户名 hyan 成立）
--   hyan1 / changyears45  -> 45377e13006cdc5a99513126b86a54e4
--   hyan2 / changyears45  -> 22adb69ee809438ca2efbe47b2fff5e7
--   admin / root           -> 5709aa2e0d3394e4553eb846b765c0e7（站点管理员 is_site_admin=true）
--
-- 业务数据：板块 t 由 hyan1 创建；帖子 t1/c1（hyan1）、t2/c2（hyan2）均在 t 板；附示例评论便于联调。

INSERT INTO "user" (user_id, username, password, is_site_admin) VALUES
    (10001, 'hyan', 'c6f3780db708c2ad91a343b14c4e3b77', FALSE),
    (10002, 'hyan1', '45377e13006cdc5a99513126b86a54e4', FALSE),
    (10003, 'hyan2', '22adb69ee809438ca2efbe47b2fff5e7', FALSE),
    (10004, 'admin', '5709aa2e0d3394e4553eb846b765c0e7', TRUE);

INSERT INTO "board" (slug, name, description, is_system_sink, created_by) VALUES
    ('_archived', '已归档', '原板块删除后，帖子统一归并到此板块。', TRUE, NULL),
    ('general', '综合', '默认讨论板块。', FALSE, NULL),
    ('t', 'T 板块', 'hyan1 创建的测试板块。', FALSE, 10002);

INSERT INTO "post" (board_id, title, content, author_id)
SELECT b.id, 't1', 'c1', 10002 FROM "board" b WHERE b.slug = 't'
UNION ALL
SELECT b.id, 't2', 'c2', 10003 FROM "board" b WHERE b.slug = 't';

-- 示例评论：hyan 在帖子 t1 下顶楼；hyan1 回复 hyan（楼中楼）
INSERT INTO "comment" (post_id, author_id, parent_id, content)
SELECT p.id, 10001, NULL, 'hyan 对 t1 的顶楼评论'
FROM "post" p
JOIN "board" b ON b.id = p.board_id
WHERE b.slug = 't' AND p.title = 't1'
LIMIT 1;

INSERT INTO "comment" (post_id, author_id, parent_id, content)
SELECT p.id, 10002, c.id, 'hyan1 回复 hyan'
FROM "post" p
JOIN "board" b ON b.id = p.board_id
JOIN "comment" c ON c.post_id = p.id AND c.parent_id IS NULL AND c.author_id = 10001
WHERE b.slug = 't' AND p.title = 't1'
LIMIT 1;

-- 全站默认标签
INSERT INTO "tag" (slug, name, description) VALUES
    ('discussion', '讨论', '普通讨论类内容'),
    ('question', '提问', '问题求助'),
    ('show', '展示', '作品展示')
ON CONFLICT (slug) DO NOTHING;

-- 给种子帖子打标签：
-- t1 -> discussion, question
-- t2 -> discussion, show
INSERT INTO "post_tag" (post_id, tag_id)
SELECT p.id, t.id
FROM "post" p, "tag" t
WHERE p.title = 't1' AND t.slug IN ('discussion', 'question')
UNION ALL
SELECT p.id, t.id
FROM "post" p, "tag" t
WHERE p.title = 't2' AND t.slug IN ('discussion', 'show')
ON CONFLICT (post_id, tag_id) DO NOTHING;

-- hyan1 收藏综合板与 T 板（联调收藏列表 / is_favorited）
INSERT INTO board_favorite (user_id, board_id, create_time)
SELECT 10002, b.id, NOW()
FROM board b
WHERE b.slug IN ('general', 't')
ON CONFLICT (user_id, board_id) DO NOTHING;

-- hyan1 收藏帖子 t1（联调帖子收藏）
INSERT INTO post_favorite (user_id, post_id, create_time)
SELECT 10002, p.id, NOW()
FROM "post" p
WHERE p.title = 't1'
ON CONFLICT (user_id, post_id) DO NOTHING;

-- =========================
-- 排序联调样本：100 条测试帖子
-- 设计目标：
-- 1) 板块分布：general(60) + t(40)
-- 2) 时间分布：最近 2 小时到约 8 天前，便于观察 new/hot
-- 3) 分数分布：包含高分、低分与负分，便于观察 top/hot
-- 4) 投票分布：补齐 post_vote，再把 score 同步为净票数
-- =========================
INSERT INTO "post" (board_id, title, content, author_id, score, create_time, update_time)
SELECT
    b.id,
    format('sort-lab-%s', gs.n),
    format(
        '排序实验样本 #%s。board=%s, author=%s, seed_score=%s, age_hour≈%s',
        gs.n,
        b.slug,
        CASE (gs.n % 4) WHEN 0 THEN 'hyan(10001)' WHEN 1 THEN 'hyan1(10002)' WHEN 2 THEN 'hyan2(10003)' ELSE 'admin(10004)' END,
        ((gs.n * 7) % 120) - 35,
        ((gs.n * 11) % 192) + 1
    ),
    CASE (gs.n % 4)
        WHEN 0 THEN 10001
        WHEN 1 THEN 10002
        WHEN 2 THEN 10003
        ELSE 10004
    END AS author_id,
    ((gs.n * 7) % 120) - 35 AS score,
    NOW() - make_interval(hours => ((gs.n * 11) % 192) + 1),
    NOW() - make_interval(hours => ((gs.n * 11) % 192) + 1)
FROM generate_series(1, 100) AS gs(n)
JOIN "board" b ON b.slug = CASE WHEN gs.n <= 60 THEN 'general' ELSE 't' END;

-- 给这 100 条样本帖生成投票（每帖 1~4 票，含上下票）
-- 票型规则（按 n%5 分组）：
--   0: + + + +   (净 +4)
--   1: + + + -   (净 +2)
--   2: + + - -   (净  0)
--   3: + - - -   (净 -2)
--   4: - - - -   (净 -4)
INSERT INTO "post_vote" (post_id, user_id, value, create_time, update_time)
SELECT
    p.id,
    v.user_id,
    v.value,
    p.create_time + make_interval(mins => v.seq * 3),
    p.create_time + make_interval(mins => v.seq * 3)
FROM "post" p
JOIN LATERAL (
    SELECT * FROM (
        VALUES
            (1, 10001::bigint, CASE (split_part(p.title, '-', 3)::int % 5) WHEN 0 THEN 1 WHEN 1 THEN 1 WHEN 2 THEN 1 WHEN 3 THEN 1 ELSE -1 END),
            (2, 10002::bigint, CASE (split_part(p.title, '-', 3)::int % 5) WHEN 0 THEN 1 WHEN 1 THEN 1 WHEN 2 THEN 1 WHEN 3 THEN -1 ELSE -1 END),
            (3, 10003::bigint, CASE (split_part(p.title, '-', 3)::int % 5) WHEN 0 THEN 1 WHEN 1 THEN 1 WHEN 2 THEN -1 WHEN 3 THEN -1 ELSE -1 END),
            (4, 10004::bigint, CASE (split_part(p.title, '-', 3)::int % 5) WHEN 0 THEN 1 WHEN 1 THEN -1 WHEN 2 THEN -1 WHEN 3 THEN -1 ELSE -1 END)
    ) AS t(seq, user_id, value)
) AS v ON TRUE
WHERE p.title LIKE 'sort-lab-%'
ON CONFLICT (post_id, user_id) DO NOTHING;

-- 将样本帖 score 回写为净票数，确保 score 与 post_vote 对齐
UPDATE "post" p
SET score = pv.net_score,
    update_time = NOW()
FROM (
    SELECT post_id, SUM(value)::bigint AS net_score
    FROM "post_vote"
    GROUP BY post_id
) pv
WHERE p.id = pv.post_id
  AND p.title LIKE 'sort-lab-%';