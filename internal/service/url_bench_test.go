package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"go.uber.org/zap"
)

func setupBench(b *testing.B) (*URLService, repository.URLRepository, func()) {
	b.Helper()

	tmpfile := fmt.Sprintf("/tmp/bench-%d.json", time.Now().UnixNano())
	repo, err := repository.NewMemoryStorage(tmpfile)
	if err != nil {
		b.Fatal(err)
	}

	svc := NewURLService(
		repo,
		"http://localhost:8080",
		1000,
		4,
		5*time.Second,
		zap.NewNop(),
	)

	cleanup := func() {
		svc.Shutdown()
	}

	return svc, repo, cleanup
}

func BenchmarkURLService_Shorten(b *testing.B) {
	svc, _, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"
	req := model.RequestShorten{URL: "https://example.com/very/long/path"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Shorten(ctx, req, userID)
	}
}

func BenchmarkURLService_BaseGet(b *testing.B) {
	svc, repo, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"
	shortID := "bench123"
	originalURL := "https://example.com/original"

	_ = repo.Save(ctx, shortID, originalURL, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.BaseGet(ctx, shortID)
	}
}

func BenchmarkURLService_ShortenBatch_10(b *testing.B) {
	benchmarkShortenBatch(b, 10)
}

func BenchmarkURLService_ShortenBatch_100(b *testing.B) {
	benchmarkShortenBatch(b, 100)
}

func BenchmarkURLService_ShortenBatch_1000(b *testing.B) {
	benchmarkShortenBatch(b, 1000)
}

func benchmarkShortenBatch(b *testing.B, batchSize int) {
	svc, _, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"

	reqs := make([]model.RequestShortenBatch, batchSize)
	for i := 0; i < batchSize; i++ {
		reqs[i] = model.RequestShortenBatch{
			CorrelationID: fmt.Sprintf("corr-%d-%d", batchSize, i),
			OriginalURL:   fmt.Sprintf("https://example.com/path/%d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ShortenBatch(ctx, reqs, userID)
	}
}

func BenchmarkURLService_GetUserUrls_10(b *testing.B) {
	benchmarkGetUserUrls(b, 10)
}

func BenchmarkURLService_GetUserUrls_100(b *testing.B) {
	benchmarkGetUserUrls(b, 100)
}

func benchmarkGetUserUrls(b *testing.B, urlCount int) {
	svc, repo, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"

	for i := 0; i < urlCount; i++ {
		shortID := fmt.Sprintf("u%d", i)
		originalURL := fmt.Sprintf("https://example.com/user/%d", i)
		_ = repo.Save(ctx, shortID, originalURL, userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetUserUrls(ctx, userID)
	}
}

func BenchmarkURLService_Concurrent_Shorten(b *testing.B) {
	svc, _, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			req := model.RequestShorten{
				URL: fmt.Sprintf("https://example.com/concurrent/%d", counter),
			}
			counter++
			_, _ = svc.Shorten(ctx, req, userID)
		}
	})
}

func BenchmarkURLService_DeleteUserUrls(b *testing.B) {
	svc, _, cleanup := setupBench(b)
	defer cleanup()

	ctx := context.Background()
	userID := "bench-user"
	urls := []string{"abc12345", "def67890", "ghi11111", "jkl22222", "mno33333"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.DeleteUserUrls(ctx, urls, userID)
	}
}
