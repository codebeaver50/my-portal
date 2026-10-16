// Package openbd は openBD (https://openbd.jp/) からカバー画像を取得する。
// openBDは無料・認証不要のISBN検索専用APIで、NDLがサムネイルを提供しない
// 場合のフォールバックのカバー取得元として使う。
package openbd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const endpoint = "https://api.openbd.jp/v1/get"

// httpClient はopenBDが応答しない場合にsyncのバッチが止まり続けないよう、タイムアウトを設定する。
var httpClient = &http.Client{Timeout: 30 * time.Second}

type record struct {
	Summary *struct {
		ISBN  string `json:"isbn"`
		Cover string `json:"cover"`
	} `json:"summary"`
}

// FetchCovers は指定のISBNのうちopenBDにカバー画像があるものについて、
// isbn13 -> カバー画像URL のマップを返す。空の入力の場合はリクエストせずに
// 空のマップを返す。
func FetchCovers(ctx context.Context, isbn13List []string) (map[string]string, error) {
	covers := make(map[string]string)
	if len(isbn13List) == 0 {
		return covers, nil
	}

	joined := strings.Join(isbn13List, ",")

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("isbn", joined)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openBD request failed: %d %s", resp.StatusCode, resp.Status)
	}

	var records []record
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Summary == nil || r.Summary.ISBN == "" || r.Summary.Cover == "" {
			continue
		}
		covers[r.Summary.ISBN] = r.Summary.Cover
	}

	return covers, nil
}
