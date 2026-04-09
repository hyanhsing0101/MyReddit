package postgres

import (
	"errors"

	"github.com/lib/pq"
)

// Int64Array 统一封装 PG bigint[] 绑定，避免调用方散落 pq.Array。
func Int64Array(v []int64) interface{} {
	return pq.Array(v)
}

const pgCodeUniqueViolation = "23505"

// IsPGCode 判断错误是否为指定的 PostgreSQL 错误码。
func IsPGCode(err error, code string) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == code
}

// IsUniqueViolation 判断是否为唯一键冲突（23505）。
func IsUniqueViolation(err error) bool {
	return IsPGCode(err, pgCodeUniqueViolation)
}
