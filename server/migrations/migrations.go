// Package migrations はSQLマイグレーションのファイルをバイナリに埋め込む。
// デプロイ時に、マイグレーション用のCLIやSQLファイルを別に配置しなくて済むようにするため。
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
