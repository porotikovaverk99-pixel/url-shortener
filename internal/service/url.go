package service

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"go.uber.org/zap"
)

const (
	shortIDLength       = 8
	maxGenerateAttempts = 100
)

var (
	// ErrEmptyRequest возвращается при пустом запросе
	ErrEmptyRequest = errors.New("empty request")
	// ErrInvalidURL возвращается при некорректном URL
	ErrInvalidURL = errors.New("invalid URL")
	// ErrMissingCorrelationID возвращается при отсутствии correlation ID в пакетном запросе
	ErrMissingCorrelationID = errors.New("missing correlation ID")
	// ErrQueueFull возвращается когда очередь удаления переполнена
	ErrQueueFull = errors.New("delete queue is full")

	// ErrFailedToGenerateID возвращается при невозможности сгенерировать уникальный ID
	ErrFailedToGenerateID = errors.New("failed to generate unique ID")
	// ErrURLAlreadyExists возвращается когда URL уже существует
	ErrURLAlreadyExists = errors.New("URL already exists")
	// ErrIDAlreadyExists возвращается когда сгенерированный ID уже существует
	ErrIDAlreadyExists = errors.New("ID already exists")
	// ErrURLNotFound возвращается когда URL не найден
	ErrURLNotFound = errors.New("URL not found")
	// ErrURLDeleted возвращается когда URL был удален
	ErrURLDeleted = errors.New("URL deleted")

	// ErrRepository возвращается при ошибке репозитория
	ErrRepository = errors.New("repository error")
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// URLService предоставляет бизнес-логику для работы с URL.
type URLService struct {
	repo          repository.URLRepository
	baseURL       string
	deleteQueue   chan model.DeleteTask
	cancelFunc    context.CancelFunc
	workerCount   int
	wg            sync.WaitGroup
	deleteTimeout time.Duration
	log           *zap.Logger
}

// NewURLService создает новый экземпляр URLService.
// Параметры:
//   - repo - репозиторий для хранения URL
//   - baseURL - базовый URL сервиса
//   - queueSize - размер очереди удаления
//   - workerCount - количество воркеров для асинхронного удаления
//   - deleteTimeout - таймаут на операцию удаления
//   - log - логгер
func NewURLService(
	repo repository.URLRepository,
	baseURL string,
	queueSize, workerCount int,
	deleteTimeout time.Duration,
	log *zap.Logger,
) *URLService {
	deleteQueue := make(chan model.DeleteTask, queueSize)
	svc := &URLService{
		repo:          repo,
		baseURL:       baseURL,
		deleteQueue:   deleteQueue,
		workerCount:   workerCount,
		deleteTimeout: deleteTimeout,
		log:           log,
	}
	svc.startWorkers()
	return svc
}

// startWorkers запускает воркеры для асинхронного удаления URL.
func (s *URLService) startWorkers() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.deletionWorker(ctx, i)
	}

	s.log.Info("Started deletion workers",
		zap.Int("worker_count", s.workerCount),
		zap.Int("queue_size", cap(s.deleteQueue)),
		zap.Duration("delete_timeout", s.deleteTimeout),
	)
}

// deletionWorker обрабатывает задачи на удаление URL из очереди.
func (s *URLService) deletionWorker(ctx context.Context, workerID int) {
	s.log.Info("Deletion worker started", zap.Int("worker_id", workerID))
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Deletion worker stopped by context",
				zap.Int("worker_id", workerID))
			return

		case task, ok := <-s.deleteQueue:
			if !ok {
				s.log.Info("Deletion worker: queue closed",
					zap.Int("worker_id", workerID))
				return
			}

			s.log.Info("Processing delete task",
				zap.Int("worker_id", workerID),
				zap.String("user_id", task.UserID),
				zap.Int("url_count", len(task.ShortURLs)),
			)

			deleteCtx, cancel := context.WithTimeout(context.Background(), s.deleteTimeout)
			defer cancel()

			err := s.repo.MarkURLsAsDeleted(deleteCtx, task.ShortURLs, task.UserID)
			if err != nil {
				s.log.Error("Failed to delete URLs",
					zap.Int("worker_id", workerID),
					zap.String("user_id", task.UserID),
					zap.Error(err),
				)
			} else {
				s.log.Info("URLs deleted successfully",
					zap.Int("worker_id", workerID),
					zap.String("user_id", task.UserID),
					zap.Int("url_count", len(task.ShortURLs)),
				)
			}
		}
	}
}

// Shutdown останавливает сервис и ожидает завершения всех воркеров.
func (s *URLService) Shutdown() {
	s.log.Info("Shutting down URL service")

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	s.wg.Wait()

	close(s.deleteQueue)

	s.log.Info("URL service shutdown completed")
}

var charsBytes = []byte(chars)

// generateShortID генерирует случайный короткий идентификатор заданной длины.
func generateShortID(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = charsBytes[rand.Intn(len(charsBytes))]
	}
	return string(b)
}

// processURL обрабатывает URL: проверяет существование или создает новый.
func processURL(ctx context.Context, repo repository.URLRepository, url string, userID string) (string, error) {
	foundID, err := repo.FindIDByURL(ctx, url, userID)
	if err == nil {
		return foundID, ErrURLAlreadyExists
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		id := generateShortID(shortIDLength)
		err = repo.Save(ctx, id, url, userID)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, repository.ErrIDAlreadyExists) {
			return "", ErrRepository
		}
	}

	return "", ErrFailedToGenerateID
}

// Ping проверяет доступность базы данных.
func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// GetUserUrls возвращает все URL принадлежащие пользователю.
func (s *URLService) GetUserUrls(ctx context.Context, userID string) ([]model.ResponseGetUserUrls, error) {
	result, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return []model.ResponseGetUserUrls{}, nil
	}

	ressGetAll := make([]model.ResponseGetUserUrls, 0, len(result))
	for originalURL, shortURL := range result {
		ressGetAll = append(ressGetAll, model.ResponseGetUserUrls{
			ShortURL:    s.baseURL + "/" + shortURL,
			OriginalURL: originalURL,
		})
	}
	return ressGetAll, nil
}

// ShortenBatch выполняет пакетное создание коротких URL.
func (s *URLService) ShortenBatch(ctx context.Context, reqsBatch []model.RequestShortenBatch, userID string) (*model.BatchResult, error) {
	if len(reqsBatch) == 0 {
		return nil, ErrEmptyRequest
	}

	urls := make([]string, len(reqsBatch))

	for i, item := range reqsBatch {
		if item.OriginalURL == "" {
			return nil, ErrInvalidURL
		}
		if item.CorrelationID == "" {
			return nil, ErrMissingCorrelationID
		}
		urls[i] = item.OriginalURL
	}

	foundIDs, err := s.repo.FindIDByURLs(ctx, urls, userID)
	if err != nil {
		return nil, ErrRepository
	}

	ressBatch := make([]model.ResponseShortenBatch, 0, len(reqsBatch))
	batch := make([]model.BatchItem, 0, len(reqsBatch))
	createdNew := false

	for _, item := range reqsBatch {
		if foundID, ok := foundIDs[item.OriginalURL]; ok {
			ressBatch = append(ressBatch, model.ResponseShortenBatch{
				CorrelationID: item.CorrelationID,
				ShortURL:      s.baseURL + "/" + foundID,
			})
		} else {
			createdNew = true
			generatedID := generateShortID(shortIDLength)
			batch = append(batch, model.BatchItem{
				ShortURL:    generatedID,
				OriginalURL: item.OriginalURL,
			})
			ressBatch = append(ressBatch, model.ResponseShortenBatch{
				CorrelationID: item.CorrelationID,
				ShortURL:      s.baseURL + "/" + generatedID,
			})
		}
	}

	if createdNew {
		err = s.repo.SaveBatch(ctx, batch, userID)
		if err != nil {
			if errors.Is(err, repository.ErrIDAlreadyExists) {
				return nil, ErrIDAlreadyExists
			}
			return nil, ErrRepository
		}
	}

	return &model.BatchResult{
		Responses:  ressBatch,
		CreatedNew: createdNew,
	}, nil
}

// Shorten создает короткий URL для переданного URL.
func (s *URLService) Shorten(ctx context.Context, reqs model.RequestShorten, userID string) (model.ResponseShorten, error) {
	if reqs.URL == "" {
		return model.ResponseShorten{}, ErrInvalidURL
	}

	id, err := processURL(ctx, s.repo, reqs.URL, userID)

	ress := model.ResponseShorten{
		Result: s.baseURL + "/" + id,
	}

	return ress, err
}

// BaseGet возвращает оригинальный URL по короткому идентификатору.
func (s *URLService) BaseGet(ctx context.Context, shortURL string) (string, error) {
	original, err := s.repo.Get(ctx, shortURL)

	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return "", ErrURLNotFound
		} else if errors.Is(err, repository.ErrURLDeleted) {
			return "", ErrURLDeleted
		} else {
			return "", ErrRepository
		}
	}
	return original, nil
}

// BasePost создает короткий URL из текстового тела запроса.
func (s *URLService) BasePost(ctx context.Context, originalURL string, userID string) (string, error) {
	if originalURL == "" {
		return "", ErrInvalidURL
	}

	id, err := processURL(ctx, s.repo, originalURL, userID)

	return s.baseURL + "/" + id, err
}

// DeleteUserUrls добавляет URL в очередь на удаление.
func (s *URLService) DeleteUserUrls(ctx context.Context, reqs []string, userID string) error {
	if len(reqs) == 0 {
		return ErrEmptyRequest
	}

	task := model.DeleteTask{
		UserID:    userID,
		ShortURLs: reqs,
	}

	select {
	case s.deleteQueue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}
