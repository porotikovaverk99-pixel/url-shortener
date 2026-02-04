package service

import (
	"context"
	"errors"
	"log"

	"math/rand"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
)

var (
	ErrEmptyRequest         = errors.New("empty request")
	ErrInvalidURL           = errors.New("invalid URL")
	ErrMissingCorrelationID = errors.New("missing correlation ID")
	ErrQueueFull            = errors.New("delete queue is full")

	ErrFailedToGenerateID = errors.New("failed to generate unique ID")
	ErrURLAlreadyExists   = errors.New("URL already exists")
	ErrIDAlreadyExists    = errors.New("ID already exists")
	ErrURLNotFound        = errors.New("URL not found")
	ErrURLDeleted         = errors.New("URL deleted")

	ErrRepository = errors.New("repository error")
)

type URLService struct {
	repo        repository.URLRepository
	baseURL     string
	deleteQueue chan model.DeleteTask
	cancelFunc  context.CancelFunc
	workerCount int
}

func NewURLService(repo repository.URLRepository, baseURL string, queueSize, workerCount int) *URLService {
	deleteQueue := make(chan model.DeleteTask, queueSize)
	svc := &URLService{
		repo:        repo,
		baseURL:     baseURL,
		deleteQueue: deleteQueue,
		workerCount: workerCount,
	}
	svc.startWorkers()
	return svc
}

func (s *URLService) startWorkers() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	for i := 0; i < s.workerCount; i++ {
		go s.deletionWorker(ctx, i)
	}
}

func (s *URLService) deletionWorker(ctx context.Context, workerID int) {
	log.Printf("Deletion worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Deletion worker %d stopped", workerID)
			return

		case task, ok := <-s.deleteQueue:
			if !ok {
				log.Printf("Deletion worker %d: queue closed", workerID)
				return
			}

			err := s.repo.MarkURLsAsDeleted(ctx, task.ShortURLs, task.UserID)
			if err != nil {
				log.Printf("Worker %d: failed to delete URLs for user %s: %v",
					workerID, task.UserID, err)
			}
		}
	}
}

func (s *URLService) Shutdown() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	close(s.deleteQueue)
}

func generateShortID(l int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, l)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func processURL(ctx context.Context, repo repository.URLRepository, url string, userID string) (string, error) {

	foundID, err := repo.FindIDByURL(ctx, url, userID)
	if err == nil {
		return foundID, ErrURLAlreadyExists
	}

	id := generateShortID(8)
	err = repo.Save(ctx, id, url, userID)

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrIDAlreadyExists):
			return "", ErrIDAlreadyExists
		default:
			return "", ErrRepository
		}
	}

	return id, nil

}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *URLService) GetUserUrls(ctx context.Context, userID string) ([]model.ResponseGetUserUrls, error) {
	result, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
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

func (s *URLService) ShortenBatch(ctx context.Context, reqsBatch []model.RequestShortenBatch, userID string) (*model.BatchResult, error) {

	if len(reqsBatch) == 0 {
		return nil, ErrEmptyRequest
	}

	urls := make([]string, 0, len(reqsBatch))

	for _, item := range reqsBatch {
		if item.OriginalURL == "" {
			return nil, ErrInvalidURL
		}
		if item.CorrelationID == "" {
			return nil, ErrMissingCorrelationID
		}
		urls = append(urls, item.OriginalURL)
	}

	foundIDs, err := s.repo.FindIDByURLs(ctx, urls, userID)
	if err != nil {
		return nil, ErrRepository
	}

	batch := []model.BatchItem{}
	ressBatch := make([]model.ResponseShortenBatch, 0, len(reqsBatch))
	generatedIDs := make(map[string]bool)

	for _, item := range reqsBatch {
		if foundID, ok := foundIDs[item.OriginalURL]; ok {
			ressBatch = append(ressBatch, model.ResponseShortenBatch{CorrelationID: item.CorrelationID, ShortURL: s.baseURL + "/" + foundID})
		} else {
			generatedID := generateShortID(8)
			attempts := 0
			for generatedIDs[generatedID] && attempts < 100 {
				generatedID = generateShortID(8)
				attempts++
			}
			if attempts >= 100 {
				return nil, ErrFailedToGenerateID
			}
			generatedIDs[generatedID] = true
			batch = append(batch, model.BatchItem{ShortURL: generatedID, OriginalURL: item.OriginalURL})
			ressBatch = append(ressBatch, model.ResponseShortenBatch{CorrelationID: item.CorrelationID, ShortURL: s.baseURL + "/" + generatedID})
		}
	}

	createdNew := len(batch) > 0

	if createdNew {
		err = s.repo.SaveBatch(ctx, batch, userID)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrIDAlreadyExists):
				return nil, ErrIDAlreadyExists
			default:
				return nil, ErrRepository
			}
		}
	}

	return &model.BatchResult{
		Responses:  ressBatch,
		CreatedNew: createdNew,
	}, nil
}

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

func (s *URLService) BasePost(ctx context.Context, originalURL string, userID string) (string, error) {

	if originalURL == "" {
		return "", ErrInvalidURL
	}

	id, err := processURL(ctx, s.repo, originalURL, userID)

	return s.baseURL + "/" + id, err
}

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
