package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var ErrNotFound error = errors.New("not found")

var db *gorm.DB

func Init() error {
	logger := log.GlobalLogger().Named("db")
	l := &dbLogger{
		l: logger,
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=true&charset=utf8&parseTime=true", config.GConf.DbConfig.User,
		config.GConf.DbConfig.Password, config.GConf.DbConfig.Endpoint, config.GConf.DbConfig.DbName)
	ldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: l,
	})
	if err != nil {
		return err
	}
	db = ldb

	return nil
}

type dbLogger struct {
	l *log.Logger
}

func (l *dbLogger) LogMode(gorm_logger.LogLevel) gorm_logger.Interface {

	return l
}
func (l *dbLogger) Info(ctx context.Context, s string, args ...interface{}) {
	l.l.Infof(s, args...)
}
func (l *dbLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.l.Warnf(s, args...)
}
func (l *dbLogger) Error(ctx context.Context, s string, args ...interface{}) {
	l.l.Errorf(s, args...)
}
func (l *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	return
}
