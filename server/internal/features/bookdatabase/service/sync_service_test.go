package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/ndl"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// fakeNDLClient は NDLClient をネットワークアクセスなしで差し替える。
// SearchSru の結果はstartRecordごとに登録し、Syncが実際に発行した
// startRecordの並びをcallsで検証できるようにする。
type fakeNDLClient struct {
	calls   []fakeNDLCall
	byStart map[int]*ndl.SearchResult
	err     error
}

type fakeNDLCall struct {
	query          string
	startRecord    int
	maximumRecords int
}

func (f *fakeNDLClient) SearchSru(query string, startRecord, maximumRecords int) (*ndl.SearchResult, error) {
	f.calls = append(f.calls, fakeNDLCall{query, startRecord, maximumRecords})
	if f.err != nil {
		return nil, f.err
	}
	if result, ok := f.byStart[startRecord]; ok {
		return result, nil
	}
	return &ndl.SearchResult{}, nil
}

// fakeOpenBDClient は OpenBDClient をネットワークアクセスなしで差し替える。
type fakeOpenBDClient struct {
	calls  [][]string
	covers map[string]string
	err    error
}

func (f *fakeOpenBDClient) FetchCovers(isbn13List []string) (map[string]string, error) {
	f.calls = append(f.calls, isbn13List)
	if f.err != nil {
		return nil, f.err
	}
	if f.covers == nil {
		return map[string]string{}, nil
	}
	return f.covers, nil
}

type syncTestDeps struct {
	db            *gorm.DB
	bookRepo      *repository.BookRepository
	publisherRepo *repository.PublisherRepository
	authorRepo    *repository.AuthorRepository
}

func newSyncTestDeps(t *testing.T) syncTestDeps {
	t.Helper()
	db := newTestDB(t)
	return syncTestDeps{
		db:            db,
		bookRepo:      repository.NewBookRepository(db),
		publisherRepo: repository.NewPublisherRepository(db),
		authorRepo:    repository.NewAuthorRepository(db),
	}
}

// rdfRecord はNDL Search SRUのdcndl_v3レコード1件分の最小限のRDF/XMLを組み立てる。
// creators/publisherName/thumbnailURLが空の場合、対応する要素は出力しない。
func rdfRecord(ndlBibID, isbn13, title string, creators []string, publisherName, issued, thumbnailURL string) string {
	var creatorsXML strings.Builder
	for _, name := range creators {
		fmt.Fprintf(&creatorsXML, `<dcterms:creator><foaf:Agent><foaf:name>%s</foaf:name><dcndl:role>著</dcndl:role></foaf:Agent></dcterms:creator>`, name)
	}

	var isbnXML string
	if isbn13 != "" {
		isbnXML = fmt.Sprintf(`<dcterms:identifier rdf:datatype="http://ndl.go.jp/dcndl/terms/ISBN">%s</dcterms:identifier>`, isbn13)
	}

	var publisherXML string
	if publisherName != "" {
		publisherXML = fmt.Sprintf(`<dcterms:publisher><foaf:Agent><foaf:name>%s</foaf:name></foaf:Agent></dcterms:publisher>`, publisherName)
	}

	var itemXML string
	if thumbnailURL != "" {
		itemXML = fmt.Sprintf(`<dcndl:Item rdf:about="urn:item:%s"><foaf:thumbnail rdf:resource="%s"/></dcndl:Item>`, ndlBibID, thumbnailURL)
	}

	return fmt.Sprintf(`<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns:dc="http://purl.org/dc/elements/1.1/"
		xmlns:dcterms="http://purl.org/dc/terms/"
		xmlns:dcndl="http://ndl.go.jp/dcndl/terms/"
		xmlns:foaf="http://xmlns.com/foaf/0.1/">
		<dcndl:BibResource rdf:about="urn:bib:%s">
			<dcterms:title>%s</dcterms:title>
			%s
			%s
			<dcterms:issued>%s</dcterms:issued>
			<dcterms:identifier rdf:datatype="http://ndl.go.jp/dcndl/terms/NDLBibID">%s</dcterms:identifier>
			%s
		</dcndl:BibResource>
		%s
	</rdf:RDF>`, ndlBibID, title, creatorsXML.String(), publisherXML, issued, ndlBibID, isbnXML, itemXML)
}

const unparseableRecord = `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"></rdf:RDF>`

func TestSyncService_Sync_CreatesNewBookWithAuthorsAndPublisher(t *testing.T) {
	deps := newSyncTestDeps(t)

	record := rdfRecord("000000001", "9784297012345", "実践Goプログラミング",
		[]string{"山田太郎", "山田太郎"}, "技術評論社", "2026-08", "https://example.com/cover.jpg")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NumberOfRecords: 1, NextRecordPosition: 0, Records: []string{record}},
	}}
	openBDClient := &fakeOpenBDClient{}

	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, openBDClient)
	summary, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Fetched != 1 || summary.Created != 1 || summary.Updated != 0 || summary.Skipped != 0 {
		t.Errorf("summary = %+v", summary)
	}

	var books []models.Book
	if err := deps.db.Preload("Publisher").Preload("BookAuthors.Author").Find(&books).Error; err != nil {
		t.Fatalf("failed to load books: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("len(books) = %d, want 1", len(books))
	}
	book := books[0]
	if book.Title != "実践Goプログラミング" {
		t.Errorf("Title = %q", book.Title)
	}
	if book.Publisher == nil || book.Publisher.Name != "技術評論社" {
		t.Errorf("Publisher = %+v", book.Publisher)
	}
	if len(book.BookAuthors) != 1 {
		t.Errorf("len(BookAuthors) = %d, want 1 (duplicate creator name should be deduped)", len(book.BookAuthors))
	} else if book.BookAuthors[0].Author == nil || book.BookAuthors[0].Author.Name != "山田太郎" {
		t.Errorf("BookAuthors[0].Author = %+v", book.BookAuthors[0].Author)
	}
	if book.CoverImageURL == nil || *book.CoverImageURL != "https://example.com/cover.jpg" {
		t.Errorf("CoverImageURL = %v, want NDL thumbnail", book.CoverImageURL)
	}

	if len(openBDClient.calls) != 1 || len(openBDClient.calls[0]) != 0 {
		t.Errorf("openBD FetchCovers calls = %v, want a single call with no ISBNs (NDL cover already present)", openBDClient.calls)
	}
}

func TestSyncService_Sync_FallsBackToOpenBDCoverWhenNDLMissing(t *testing.T) {
	deps := newSyncTestDeps(t)

	const isbn = "9784297099999"
	record := rdfRecord("000000002", isbn, "OpenBDカバーの本", nil, "", "2026-08", "")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{record}},
	}}
	wantCoverURL := "https://openbd.example.com/cover.jpg"
	openBDClient := &fakeOpenBDClient{covers: map[string]string{isbn: wantCoverURL}}

	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, openBDClient)
	summary, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Created != 1 {
		t.Errorf("summary = %+v", summary)
	}

	if len(openBDClient.calls) != 1 || len(openBDClient.calls[0]) != 1 || openBDClient.calls[0][0] != isbn {
		t.Fatalf("openBD FetchCovers calls = %v, want a single call with [%q]", openBDClient.calls, isbn)
	}

	var book models.Book
	if err := deps.db.First(&book).Error; err != nil {
		t.Fatalf("failed to load book: %v", err)
	}
	if book.CoverImageURL == nil || *book.CoverImageURL != wantCoverURL {
		t.Errorf("CoverImageURL = %v, want %q", book.CoverImageURL, wantCoverURL)
	}
}

func TestSyncService_Sync_SkipsUnparseableRecords(t *testing.T) {
	deps := newSyncTestDeps(t)

	valid := rdfRecord("000000003", "", "有効なレコード", []string{"著者A"}, "", "2026-08", "")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{unparseableRecord, valid}},
	}}

	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{})
	summary, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Fetched != 2 || summary.Skipped != 1 || summary.Created != 1 {
		t.Errorf("summary = %+v", summary)
	}
}

func TestSyncService_Sync_UpdatesExistingBookOnRepeatSyncOfSameNDLBibID(t *testing.T) {
	deps := newSyncTestDeps(t)

	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{
			rdfRecord("000000004", "9784297011111", "初版タイトル", []string{"著者A"}, "", "2026-08", ""),
		}},
	}}
	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{})

	summary1, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("first Sync() error = %v", err)
	}
	if summary1.Created != 1 || summary1.Updated != 0 {
		t.Fatalf("first summary = %+v", summary1)
	}

	ndlClient.byStart = map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{
			rdfRecord("000000004", "9784297011111", "改訂版タイトル", []string{"著者A"}, "", "2026-09", ""),
		}},
	}
	summary2, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}
	if summary2.Created != 0 || summary2.Updated != 1 {
		t.Errorf("second summary = %+v", summary2)
	}

	var books []models.Book
	if err := deps.db.Find(&books).Error; err != nil {
		t.Fatalf("failed to load books: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("len(books) = %d, want 1 (should update in place, not duplicate)", len(books))
	}
	if books[0].Title != "改訂版タイトル" {
		t.Errorf("Title = %q, want 改訂版タイトル", books[0].Title)
	}

	var sources []models.BookSource
	if err := deps.db.Find(&sources).Error; err != nil {
		t.Fatalf("failed to load sources: %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("len(sources) = %d, want 1 (should update in place, not duplicate)", len(sources))
	}
}

func TestSyncService_Sync_MatchesExistingBookByISBNWhenSourceMissing(t *testing.T) {
	deps := newSyncTestDeps(t)

	const isbn = "9784297022222"
	existing := &models.Book{Title: "openBD経由で先に登録された本", ISBN13: strPtr(isbn)}
	if err := deps.db.Create(existing).Error; err != nil {
		t.Fatalf("failed to seed existing book: %v", err)
	}

	record := rdfRecord("000000007", isbn, "NDLで補完されたタイトル", []string{"著者B"}, "", "2026-08", "")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{record}},
	}}
	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{})

	summary, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Created != 0 || summary.Updated != 1 {
		t.Errorf("summary = %+v", summary)
	}

	var books []models.Book
	if err := deps.db.Find(&books).Error; err != nil {
		t.Fatalf("failed to load books: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("len(books) = %d, want 1 (should update the existing ISBN match, not create a new row)", len(books))
	}
	if books[0].ID != existing.ID {
		t.Errorf("book ID = %d, want %d (existing row)", books[0].ID, existing.ID)
	}
	if books[0].Title != "NDLで補完されたタイトル" {
		t.Errorf("Title = %q", books[0].Title)
	}

	var sources []models.BookSource
	if err := deps.db.Find(&sources).Error; err != nil {
		t.Fatalf("failed to load sources: %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("len(sources) = %d, want 1 (a new source record for this NDLBibID)", len(sources))
	}
}

func TestSyncService_Sync_PaginatesUntilNextRecordPositionExceedsLimit(t *testing.T) {
	deps := newSyncTestDeps(t)

	recordA := rdfRecord("000000005", "", "ページ1の本", nil, "", "2026-08", "")
	recordB := rdfRecord("000000006", "", "ページ2の本", nil, "", "2026-08", "")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1:   {NumberOfRecords: 2, NextRecordPosition: 201, Records: []string{recordA}},
		201: {NumberOfRecords: 2, NextRecordPosition: 0, Records: []string{recordB}},
	}}
	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{})

	summary, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Fetched != 2 || summary.Created != 2 {
		t.Errorf("summary = %+v", summary)
	}

	if len(ndlClient.calls) != 2 {
		t.Fatalf("SearchSru called %d times, want 2", len(ndlClient.calls))
	}
	if ndlClient.calls[0].startRecord != 1 || ndlClient.calls[1].startRecord != 201 {
		t.Errorf("startRecord calls = %+v", ndlClient.calls)
	}
}

func TestSyncService_Sync_PropagatesNDLClientError(t *testing.T) {
	deps := newSyncTestDeps(t)

	wantErr := errors.New("ndl unavailable")
	ndlClient := &fakeNDLClient{err: wantErr}
	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{})

	if _, err := svc.Sync(context.Background()); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestSyncService_Sync_PropagatesOpenBDClientError(t *testing.T) {
	deps := newSyncTestDeps(t)

	// カバー画像なしのレコード: openBDへのカバー検索が発生する。
	record := rdfRecord("000000008", "9784297033333", "本", nil, "", "2026-08", "")
	ndlClient := &fakeNDLClient{byStart: map[int]*ndl.SearchResult{
		1: {NextRecordPosition: 0, Records: []string{record}},
	}}
	wantErr := errors.New("openbd unavailable")
	svc := NewSyncService(deps.db, deps.bookRepo, deps.publisherRepo, deps.authorRepo, ndlClient, &fakeOpenBDClient{err: wantErr})

	if _, err := svc.Sync(context.Background()); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func strPtr(s string) *string { return &s }
