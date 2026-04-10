package postgres

import "strconv"

// PostReader 帖子列表/详情读权限：私有板可见性 + 封帖可见性（未软删由调用方与片段一并拼接）。
type PostReader struct {
	UserID    *int64
	SiteAdmin bool
}

func postReaderUID(r *PostReader) interface{} {
	if r != nil && r.UserID != nil {
		return *r.UserID
	}
	return nil
}

func postReaderAdmin(r *PostReader) bool {
	return r != nil && r.SiteAdmin
}

// sqlPostBoardReadable 需与同语句中的占位符 $uidIdx、$adminIdx 一致（用于动态拼 WHERE）。
func sqlPostBoardReadable(uidIdx, adminIdx int) string {
	u := strconv.Itoa(uidIdx)
	a := strconv.Itoa(adminIdx)
	return `(
	b.visibility = 'public'
	OR b.is_system_sink = TRUE
	OR ($` + u + `::bigint IS NOT NULL AND EXISTS (
		SELECT 1 FROM board_favorite bf WHERE bf.board_id = p.board_id AND bf.user_id = $` + u + `))
	OR ($` + u + `::bigint IS NOT NULL AND EXISTS (
		SELECT 1 FROM board_moderator bm WHERE bm.board_id = p.board_id AND bm.user_id = $` + u + `))
	OR $` + a + ` = TRUE
)`
}

func sqlPostSealReadable(uidIdx, adminIdx int) string {
	u := strconv.Itoa(uidIdx)
	a := strconv.Itoa(adminIdx)
	return `(
	p.sealed_at IS NULL
	OR ($` + u + `::bigint IS NOT NULL AND p.author_id IS NOT NULL AND p.author_id = $` + u + `)
	OR ($` + u + `::bigint IS NOT NULL AND EXISTS (
		SELECT 1 FROM board_moderator bm2 WHERE bm2.board_id = p.board_id AND bm2.user_id = $` + u + `))
	OR $` + a + ` = TRUE
)`
}
