package postgres

import (
	"errors"
	"myreddit/models"
	"strconv"
	"strings"
)

var ErrorTagNotExist = errors.New("tag not exist")
// ListTags 分页列出全站标签
func ListTags(limit, offset int) ([]models.Tag, error) {
	list := make([]models.Tag, 0)
	sqlStr := `
SELECT id, slug, name, COALESCE(description, '') AS description, create_time, update_time
FROM "tag"
ORDER BY id ASC
LIMIT $1 OFFSET $2;
`
	err := db.Select(&list, sqlStr, limit, offset)
	return list, err
}

func CountTags() (int64, error) {
	var total int64
	err := db.Get(&total, `SELECT COUNT(*) FROM "tag";`)
	return total, err
}

// GetTagsByPostID 获取帖子绑定标签（按 tag.id 升序）
func GetTagsByPostID(postID int64) ([]models.Tag, error) {
	list := make([]models.Tag, 0)
	sqlStr := `
SELECT t.id, t.slug, t.name, COALESCE(t.description, '') AS description, t.create_time, t.update_time
FROM "post_tag" pt
JOIN "tag" t ON t.id = pt.tag_id
WHERE pt.post_id = $1
ORDER BY t.id ASC;
`
	err := db.Select(&list, sqlStr, postID)
	return list, err
}

// ValidateTagIDs 去重并校验 tag_id 全都存在
// 返回值是去重后的 id 列表（保持升序，不保留原输入顺序）
func ValidateTagIDs(tagIDs []int64) ([]int64, error) {
	if len(tagIDs) == 0 {
		return []int64{}, nil
	}

	// 去重 + 过滤非法
	uniq := make(map[int64]struct{}, len(tagIDs))
	clean := make([]int64, 0, len(tagIDs))
	for _, id := range tagIDs {
		if id < 1 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		clean = append(clean, id)
	}
	if len(clean) == 0 {
		return []int64{}, nil
	}

	// 构造 IN (...) 参数
	placeholders := make([]string, 0, len(clean))
	args := make([]interface{}, 0, len(clean))
	for i, id := range clean {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
		args = append(args, id)
	}

	var count int
	sqlStr := `SELECT COUNT(*) FROM "tag" WHERE id IN (` + strings.Join(placeholders, ",") + `);`
	if err := db.Get(&count, sqlStr, args...); err != nil {
		return nil, err
	}
	if count != len(clean) {
		return nil, ErrorTagNotExist
	}
	return clean, nil
}

// ReplacePostTags 覆盖帖子标签（先删后插）
func ReplacePostTags(postID int64, tagIDs []int64) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.Exec(`DELETE FROM "post_tag" WHERE post_id = $1;`, postID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err = tx.Exec(
			`INSERT INTO "post_tag" (post_id, tag_id) VALUES ($1, $2) ON CONFLICT (post_id, tag_id) DO NOTHING;`,
			postID, tagID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}