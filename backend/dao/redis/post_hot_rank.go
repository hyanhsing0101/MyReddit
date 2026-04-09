package redis

import (
	"fmt"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis"
)

const (
	postHotRankKey       = "post:rank:hot:all"
	postHotRankTTL       = 2 * time.Minute
	postHotRankTTLJitter = 30 * time.Second
)

// PostHotRankKey 暴露给 logic 层（后续可扩展 board 维度）
func PostHotRankKey() string {
	return postHotRankKey
}

// GetHotPostIDs 从 ZSET 取某页帖子 id（按 score 高到低）。
func GetHotPostIDs(page, pageSize int) ([]int64, error) {
	if page < 1 || pageSize < 1 {
		return []int64{}, nil
	}
	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1

	members, err := rdb.ZRevRange(postHotRankKey, start, stop).Result()
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

// SetHotPostScores 批量写入 ZSET，并刷新 TTL（带轻微抖动）。
func SetHotPostScores(scores map[int64]float64) error {
	if len(scores) == 0 {
		return nil
	}
	zs := make([]goredis.Z, 0, len(scores))
	for id, score := range scores {
		zs = append(zs, goredis.Z{Score: score, Member: fmt.Sprintf("%d", id)})
	}
	if err := rdb.ZAdd(postHotRankKey, zs...).Err(); err != nil {
		return err
	}
	ttl := postHotRankTTL + time.Duration(time.Now().UnixNano()%int64(postHotRankTTLJitter))
	return rdb.Expire(postHotRankKey, ttl).Err()
}

// CountHotPosts 返回当前缓存中可分页总量（非权威值）。
func CountHotPosts() (int64, error) {
	return rdb.ZCard(postHotRankKey).Result()
}

// UpsertHotPost 更新/写入单个帖子的热度分，并刷新 TTL。
func UpsertHotPost(postID int64, score float64) error {
	if err := rdb.ZAdd(postHotRankKey, goredis.Z{
		Score:  score,
		Member: fmt.Sprintf("%d", postID),
	}).Err(); err != nil {
		return err
	}
	ttl := postHotRankTTL + time.Duration(time.Now().UnixNano()%int64(postHotRankTTLJitter))
	return rdb.Expire(postHotRankKey, ttl).Err()
}

// RemoveHotPost 从热榜中移除单个帖子（如软删）。
func RemoveHotPost(postID int64) error {
	return rdb.ZRem(postHotRankKey, fmt.Sprintf("%d", postID)).Err()
}
