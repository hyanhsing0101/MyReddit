package logic

import "myreddit/dao/postgres"

// postReader 根据可选浏览者构造帖子读权限上下文（含站主绕过）。
func postReader(viewerID *int64) (*postgres.PostReader, error) {
	if viewerID == nil {
		return &postgres.PostReader{}, nil
	}
	admin, err := postgres.IsSiteAdmin(*viewerID)
	if err != nil {
		return nil, err
	}
	uid := *viewerID
	return &postgres.PostReader{UserID: &uid, SiteAdmin: admin}, nil
}
