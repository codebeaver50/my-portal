// Package ndl は国立国会図書館サーチのSRU APIと通信し、返ってくる
// DC-NDL（RDF）形式の書誌レコードをパースする。
package ndl

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// 国立国会図書館サーチ 外部提供インタフェース仕様書（第1.4版）3.SRU を参照
const sruEndpoint = "https://ndlsearch.ndl.go.jp/api/sru"

// 「検索負荷回避のための制約」により startRecord + maximumRecords は 501 を超えられない
const maxRecordPosition = 500
const maxRecordLimit = maxRecordPosition + 1

// httpClient はNDLが応答しない場合にsyncのバッチが止まり続けないよう、タイムアウトを設定する。
var httpClient = &http.Client{Timeout: 30 * time.Second}

type sruRecord struct {
	RecordData string `xml:"recordData"`
}

type sruResponse struct {
	XMLName            xml.Name    `xml:"searchRetrieveResponse"`
	NumberOfRecords    int         `xml:"numberOfRecords"`
	NextRecordPosition int         `xml:"nextRecordPosition"`
	Records            []sruRecord `xml:"records>record"`
}

// SearchResult は SRU searchRetrieve結果の1ページ分。
type SearchResult struct {
	NumberOfRecords    int
	NextRecordPosition int
	Records            []string
}

// BuildMonthlyNdcQuery は指定のNDC分類で、指定の"YYYY-MM"月に発行された
// 書籍を対象とするSRU CQLクエリを組み立てる。
func BuildMonthlyNdcQuery(ndc string, yearMonth string) (string, error) {
	var year, month int
	if _, err := fmt.Sscanf(yearMonth, "%d-%d", &year, &month); err != nil {
		return "", fmt.Errorf("invalid yearMonth %q: %w", yearMonth, err)
	}
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	from := fmt.Sprintf("%s-01", yearMonth)
	until := fmt.Sprintf("%s-%02d", yearMonth, lastDay)
	return fmt.Sprintf(`ndc="%s" and from="%s" and until="%s"`, ndc, from, until), nil
}

// SearchSru は searchRetrieve リクエストを1回発行する。
func SearchSru(ctx context.Context, query string, startRecord, maximumRecords int) (*SearchResult, error) {
	if maxAllowed := maxRecordLimit - startRecord; maximumRecords > maxAllowed {
		maximumRecords = maxAllowed
	}

	u, err := url.Parse(sruEndpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("operation", "searchRetrieve")
	q.Set("version", "1.2")
	q.Set("recordSchema", "dcndl_v3")
	q.Set("recordPacking", "string")
	q.Set("query", query)
	q.Set("startRecord", strconv.Itoa(startRecord))
	q.Set("maximumRecords", strconv.Itoa(maximumRecords))
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
		return nil, fmt.Errorf("NDL Search SRU request failed: %d %s", resp.StatusCode, resp.Status)
	}

	var parsed sruResponse
	if err := xml.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("unexpected SRU response: %w", err)
	}

	records := make([]string, len(parsed.Records))
	for i, r := range parsed.Records {
		records[i] = r.RecordData
	}

	return &SearchResult{
		NumberOfRecords:    parsed.NumberOfRecords,
		NextRecordPosition: parsed.NextRecordPosition,
		Records:            records,
	}, nil
}

// IsWithinRecordLimit は、nextRecordPositionがAPIが次のstartRecordとして
// 実際に受け付けるレコードを指しているかを判定する。
func IsWithinRecordLimit(nextRecordPosition int) bool {
	return nextRecordPosition > 0 && nextRecordPosition <= maxRecordPosition
}
