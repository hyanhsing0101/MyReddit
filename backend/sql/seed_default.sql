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