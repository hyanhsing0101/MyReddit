package redis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func refreshTokenKey(jti string) string {
	return fmt.Sprintf("refresh_token:%s", jti)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SaveRefreshToken 保存 refresh token 的 hash，并设置 TTL
func SaveRefreshToken(jti string, refreshToken string, ttl time.Duration) error {
	key := refreshTokenKey(jti)
	return rdb.Set(key, hashToken(refreshToken), ttl).Err()
}

// VerifyRefreshToken 判断：jti 存在且 hash 匹配
func VerifyRefreshToken(jti string, refreshToken string) (bool, error) {
	key := refreshTokenKey(jti)
	stored, err := rdb.Get(key).Result()
	if err != nil {
		if IsNil(err) {
			return false, nil
		}
		return false, err
	}
	return stored == hashToken(refreshToken), nil
}

func DeleteRefreshToken(jti string) error {
	key := refreshTokenKey(jti)
	return rdb.Del(key).Err()
}
