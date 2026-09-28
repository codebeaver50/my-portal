package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/ndl"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// NDC 007: 情報科学（IT関連書籍とみなす分類）
const itNdc = "007"
const syncPageSize = 200

// SyncService は GET/POST /api/books/sync バッチを実装する。
type SyncService interface {
	Sync(ctx context.Context) (*dto.SyncSummaryResponse, error)
}

type syncService struct {
	db            *gorm.DB
	bookRepo      *repository.BookRepository
	publisherRepo *repository.PublisherRepository
	authorRepo    *repository.AuthorRepository
	ndlClient     NDLClient
	openBDClient  OpenBDClient
}

// NewSyncService は SyncService を作成する。dbは各レコードのupsertを
// 複数repositoryにまたがる1つのトランザクションとして実行するために使う。
// ndlClient/openBDClientを差し替えれば、実際のネットワークアクセスなしで
// Syncをテストできる。
func NewSyncService(
	db *gorm.DB,
	bookRepo *repository.BookRepository,
	publisherRepo *repository.PublisherRepository,
	authorRepo *repository.AuthorRepository,
	ndlClient NDLClient,
	openBDClient OpenBDClient,
) SyncService {
	return &syncService{
		db:            db,
		bookRepo:      bookRepo,
		publisherRepo: publisherRepo,
		authorRepo:    authorRepo,
		ndlClient:     ndlClient,
		openBDClient:  openBDClient,
	}
}

// Sync は今月NDLが公開したNDC 007（情報科学）分類のレコードをすべて取得し、
// カバー画像を解決し（NDLを優先、なければopenBDにフォールバック）、各件を
// DBにupsertする。
func (s *syncService) Sync(ctx context.Context) (*dto.SyncSummaryResponse, error) {
	now := time.Now()
	yearMonth := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	query, err := ndl.BuildMonthlyNdcQuery(itNdc, yearMonth)
	if err != nil {
		return nil, err
	}

	summary := &dto.SyncSummaryResponse{YearMonth: yearMonth, NDC: itNdc}
	startRecord := 1

	for {
		result, err := s.ndlClient.SearchSru(query, startRecord, syncPageSize)
		if err != nil {
			return nil, err
		}
		summary.NumberOfRecords = result.NumberOfRecords
		summary.Fetched += len(result.Records)

		var parsedRecords []*ndl.ParsedRecord
		for _, rawXML := range result.Records {
			parsed, err := ndl.ParseDcndlRecord(rawXML)
			if err != nil || parsed == nil {
				summary.Skipped++
				continue
			}
			parsedRecords = append(parsedRecords, parsed)
		}

		isbnsMissingCover := uniqueISBNsMissingCover(parsedRecords)
		openBDCovers, err := s.openBDClient.FetchCovers(isbnsMissingCover)
		if err != nil {
			return nil, err
		}

		publisherIDs, err := resolvePublisherIDs(ctx, s.publisherRepo, publisherNames(parsedRecords))
		if err != nil {
			return nil, err
		}
		authorIDs, err := resolveAuthorIDs(ctx, s.authorRepo, authorNames(parsedRecords))
		if err != nil {
			return nil, err
		}

		for _, parsed := range parsedRecords {
			coverImageURL := parsed.CoverImageURL
			if coverImageURL == nil && parsed.ISBN13 != nil {
				if url, ok := openBDCovers[*parsed.ISBN13]; ok {
					coverImageURL = &url
				}
			}

			created, err := s.upsertRecord(ctx, parsed, coverImageURL, publisherIDs, authorIDs)
			if err != nil {
				return nil, err
			}
			if created {
				summary.Created++
			} else {
				summary.Updated++
			}
		}

		if !ndl.IsWithinRecordLimit(result.NextRecordPosition) {
			break
		}
		startRecord = result.NextRecordPosition
	}

	return summary, nil
}

func publisherNames(records []*ndl.ParsedRecord) []string {
	var names []string
	for _, parsed := range records {
		if parsed.PublisherName != nil && *parsed.PublisherName != "" {
			names = append(names, *parsed.PublisherName)
		}
	}
	return names
}

func authorNames(records []*ndl.ParsedRecord) []string {
	var names []string
	for _, parsed := range records {
		for _, creator := range parsed.Creators {
			names = append(names, creator.Name)
		}
	}
	return names
}

func uniqueISBNsMissingCover(records []*ndl.ParsedRecord) []string {
	seen := make(map[string]bool)
	var isbns []string
	for _, parsed := range records {
		if parsed.CoverImageURL != nil || parsed.ISBN13 == nil {
			continue
		}
		if seen[*parsed.ISBN13] {
			continue
		}
		seen[*parsed.ISBN13] = true
		isbns = append(isbns, *parsed.ISBN13)
	}
	return isbns
}

// upsertRecord はパース済みNDLレコード1件を書き込む: publisherIDs/authorIDs
// で事前解決済みの出版社・著者IDを引き当てて、書籍をcreate-or-update
// （このNDLレコードに対応する既存book_sourceを優先し、なければISBN13で照合）、
// 著者リンクを置き換え、生のソースデータを記録する。これらの書き込みはすべて
// 1つのトランザクション内で行う。書籍がcreateされたか（updateではなく）を
// 報告する。
func (s *syncService) upsertRecord(ctx context.Context, parsed *ndl.ParsedRecord, coverImageURL *string, publisherIDs, authorIDs map[string]int64) (created bool, err error) {
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bookRepo := s.bookRepo.WithTx(tx)

		rawData, err := json.Marshal(parsed)
		if err != nil {
			return err
		}
		fetchedAt := time.Now()

		var publisherID *int64
		if parsed.PublisherName != nil && *parsed.PublisherName != "" {
			if id, ok := publisherIDs[*parsed.PublisherName]; ok {
				publisherID = &id
			}
		}

		source, err := bookRepo.FindSource(ctx, models.SourceTypeNDLSearchAPI, parsed.NDLBibID)
		if err != nil {
			return err
		}

		var bookID int64

		if source != nil {
			bookID = source.BookID
			book := newBookFromParsed(bookID, parsed, publisherID, coverImageURL)
			if err := bookRepo.Update(ctx, book); err != nil {
				return err
			}

			source.RawData = rawData
			source.FetchedAt = fetchedAt
			if err := bookRepo.UpdateSource(ctx, source); err != nil {
				return err
			}
			created = false
		} else {
			existing, err := bookRepo.FindByISBN13(ctx, parsed.ISBN13)
			if err != nil {
				return err
			}

			if existing != nil {
				bookID = existing.ID
				book := newBookFromParsed(bookID, parsed, publisherID, coverImageURL)
				if err := bookRepo.Update(ctx, book); err != nil {
					return err
				}
				created = false
			} else {
				book := newBookFromParsed(0, parsed, publisherID, coverImageURL)
				if err := bookRepo.Create(ctx, book); err != nil {
					return err
				}
				bookID = book.ID
				created = true
			}

			newSource := &models.BookSource{
				SourceType: models.SourceTypeNDLSearchAPI,
				SourceID:   parsed.NDLBibID,
				BookID:     bookID,
				RawData:    rawData,
				FetchedAt:  fetchedAt,
			}
			if err := bookRepo.CreateSource(ctx, newSource); err != nil {
				return err
			}
		}

		bookAuthors := bookAuthorsForCreators(bookID, parsed.Creators, authorIDs)
		return bookRepo.ReplaceAuthors(ctx, bookID, bookAuthors)
	})

	return created, err
}

func newBookFromParsed(id int64, parsed *ndl.ParsedRecord, publisherID *int64, coverImageURL *string) *models.Book {
	return &models.Book{
		ID:            id,
		ISBN13:        parsed.ISBN13,
		Title:         parsed.Title,
		Subtitle:      parsed.Subtitle,
		PublisherID:   publisherID,
		PublishedDate: parsed.PublishedDate,
		SeriesName:    parsed.SeriesName,
		Volume:        parsed.Volume,
		Price:         parsed.Price,
		CoverImageURL: coverImageURL,
	}
}

func bookAuthorsForCreators(bookID int64, creators []ndl.ParsedCreator, authorIDs map[string]int64) []models.BookAuthor {
	linked := make(map[int64]bool)
	var bookAuthors []models.BookAuthor
	order := 0
	for _, creator := range creators {
		authorID, ok := authorIDs[creator.Name]
		if !ok || linked[authorID] {
			continue
		}
		linked[authorID] = true

		bookAuthors = append(bookAuthors, models.BookAuthor{
			BookID:   bookID,
			AuthorID: authorID,
			Role:     creator.Role,
			Order:    order,
		})
		order++
	}
	return bookAuthors
}

// nameRecord はfind-or-createのバッチ解決で扱う、名前とIDの組。
type nameRecord struct {
	Name string
	ID   int64
}

// toNameRecords はrepositoryが返すモデルのスライスを、getで名前とIDを
// 取り出しながら[]nameRecordに変換する。
func toNameRecords[T any](items []T, get func(T) (string, int64)) []nameRecord {
	records := make([]nameRecord, len(items))
	for i, item := range items {
		name, id := get(item)
		records[i] = nameRecord{Name: name, ID: id}
	}
	return records
}

// batchResolveIDs は指定の名前（重複・空文字を含みうる）ごとにIDを解決する。
// find でまとめて既存の名前を検索し、見つからなかった名前だけ create でまとめて
// 作成する。publisher/authorのfind-or-createで共通のバッチ解決アルゴリズムを
// ここに集約し、両者の差分は find/create コールバックだけに閉じ込める。
func batchResolveIDs(
	names []string,
	find func(names []string) ([]nameRecord, error),
	create func(names []string) ([]nameRecord, error),
) (map[string]int64, error) {
	unique := uniqueStrings(names)
	ids := make(map[string]int64, len(unique))
	if len(unique) == 0 {
		return ids, nil
	}

	existing, err := find(unique)
	if err != nil {
		return nil, err
	}
	for _, rec := range existing {
		ids[rec.Name] = rec.ID
	}

	var missing []string
	for _, name := range unique {
		if _, ok := ids[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return ids, nil
	}

	created, err := create(missing)
	if err != nil {
		return nil, err
	}
	for _, rec := range created {
		ids[rec.Name] = rec.ID
	}
	return ids, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	var unique []string
	for _, v := range values {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		unique = append(unique, v)
	}
	return unique
}

// resolvePublisherIDs は指定の名前ごとに出版社のIDを解決する（find-or-create
// をページ単位でバッチ化したもの）。
func resolvePublisherIDs(ctx context.Context, publisherRepo *repository.PublisherRepository, names []string) (map[string]int64, error) {
	return batchResolveIDs(names,
		func(names []string) ([]nameRecord, error) {
			publishers, err := publisherRepo.FindByNames(ctx, names)
			if err != nil {
				return nil, err
			}
			return toNameRecords(publishers, func(p models.Publisher) (string, int64) { return p.Name, p.ID }), nil
		},
		func(names []string) ([]nameRecord, error) {
			newPublishers := make([]*models.Publisher, len(names))
			for i, name := range names {
				newPublishers[i] = &models.Publisher{Name: name}
			}
			if err := publisherRepo.CreateBatch(ctx, newPublishers); err != nil {
				return nil, err
			}
			return toNameRecords(newPublishers, func(p *models.Publisher) (string, int64) { return p.Name, p.ID }), nil
		},
	)
}

// resolveAuthorIDs は指定の名前ごとに著者のIDを解決する（find-or-createを
// ページ単位でバッチ化したもの）。
func resolveAuthorIDs(ctx context.Context, authorRepo *repository.AuthorRepository, names []string) (map[string]int64, error) {
	return batchResolveIDs(names,
		func(names []string) ([]nameRecord, error) {
			authors, err := authorRepo.FindByNames(ctx, names)
			if err != nil {
				return nil, err
			}
			return toNameRecords(authors, func(a models.Author) (string, int64) { return a.Name, a.ID }), nil
		},
		func(names []string) ([]nameRecord, error) {
			newAuthors := make([]*models.Author, len(names))
			for i, name := range names {
				newAuthors[i] = &models.Author{Name: name}
			}
			if err := authorRepo.CreateBatch(ctx, newAuthors); err != nil {
				return nil, err
			}
			return toNameRecords(newAuthors, func(a *models.Author) (string, int64) { return a.Name, a.ID }), nil
		},
	)
}
