package service

import (
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/ndl"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/openbd"
)

// NDLClient は SyncService が依存するNDL Search SRU APIの操作を抽象化する。
// テストではネットワークアクセスなしのフェイクに差し替えられる。
type NDLClient interface {
	SearchSru(query string, startRecord, maximumRecords int) (*ndl.SearchResult, error)
}

// OpenBDClient は SyncService が依存するopenBDのカバー画像検索を抽象化する。
// テストではネットワークアクセスなしのフェイクに差し替えられる。
type OpenBDClient interface {
	FetchCovers(isbn13List []string) (map[string]string, error)
}

type defaultNDLClient struct{}

// NewNDLClient は実際のNDL Search SRU APIを呼び出す NDLClient を返す。
func NewNDLClient() NDLClient {
	return defaultNDLClient{}
}

func (defaultNDLClient) SearchSru(query string, startRecord, maximumRecords int) (*ndl.SearchResult, error) {
	return ndl.SearchSru(query, startRecord, maximumRecords)
}

type defaultOpenBDClient struct{}

// NewOpenBDClient は実際のopenBD APIを呼び出す OpenBDClient を返す。
func NewOpenBDClient() OpenBDClient {
	return defaultOpenBDClient{}
}

func (defaultOpenBDClient) FetchCovers(isbn13List []string) (map[string]string, error) {
	return openbd.FetchCovers(isbn13List)
}
