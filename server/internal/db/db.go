// Package db は全featureで共有するGORMのDB接続を開く。
package db

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open は "mysql://user:password@tcp(host:port)/dbname" 形式の DATABASE_URL
// （server/cmd/migrate と同じ値）でDBに接続する。
func Open(databaseURL string) (*gorm.DB, error) {
	dsn := toDSN(databaseURL)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger()})
}

// newLogger はGORMのデフォルトのロガーと同じ設定で、gorm.ErrRecordNotFound だけを
// エラーとしてログに出さないようにする。出版社・著者の名前での検索のように、
// 見つからないことが通常の結果として扱われる検索が頻繁にあるため。
func newLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
}

// toDSN は go-sql-driver が受け付けない "mysql://" を取り除き、DSNに変換する。
// DATETIME の列を time.Time として読み込むため parseTime を有効にする。また、接続の
// 文字コードを utf8mb4 に固定する（指定しないと別の文字コードで接続され、日本語等の
// ASCII以外の文字が読み書きで化けることがあるため）。
func toDSN(databaseURL string) string {
	dsn := strings.TrimPrefix(databaseURL, "mysql://")
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "parseTime=true&charset=utf8mb4"
}
