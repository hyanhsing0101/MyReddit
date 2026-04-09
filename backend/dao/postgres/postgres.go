package postgres

import (
	"fmt"
	"myreddit/settings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var db *sqlx.DB

func Init(cfg *settings.PostgresConfig) (err error) {
	ssl := cfg.SslMode
	if ssl == "" {
		ssl = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DbName,
		ssl,
	)
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		zap.L().Error("Error connecting postgres", zap.Error(err))
		return
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	return
}

func Close() {
	_ = db.Close()
}
