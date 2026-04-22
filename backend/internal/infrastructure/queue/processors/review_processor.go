package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// ReviewProcessor handles review_sync jobs (scrape app store reviews)
type ReviewProcessor struct {
	scraper     service.ReviewScraper
	reviewRepo  repository.AppReviewRepository
	appRepo     repository.AppRepository
	syncJobRepo repository.SyncJobRepository
	lockManager *queue.LockManager
	progress    *queue.ProgressTracker
}

func NewReviewProcessor(
	scraper service.ReviewScraper,
	reviewRepo repository.AppReviewRepository,
	appRepo repository.AppRepository,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *ReviewProcessor {
	return &ReviewProcessor{
		scraper:     scraper,
		reviewRepo:  reviewRepo,
		appRepo:     appRepo,
		syncJobRepo: syncJobRepo,
		lockManager: lockManager,
		progress:    progress,
	}
}

func (p *ReviewProcessor) Type() string { return entity.SyncJobTypeReviewSync }

func (p *ReviewProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	app, err := p.appRepo.FindByID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find app: %w", err)
	}

	if app.AppStoreSlug == "" {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{Message: "No app store slug configured"})
		return p.syncJobRepo.MarkCompleted(ctx, payload.JobID)
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{Message: "Scraping app store reviews..."})

	if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
		return fmt.Errorf("job cancelled")
	}

	scraped, err := p.scraper.ScrapeReviews(ctx, app.AppStoreSlug, 2)
	if err != nil {
		return fmt.Errorf("failed to scrape reviews: %w", err)
	}

	if len(scraped) == 0 {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{Message: "No reviews found"})
		return p.syncJobRepo.MarkCompleted(ctx, payload.JobID)
	}

	now := time.Now().UTC()
	var reviews []*entity.AppReview
	for _, sr := range scraped {
		if sr.Rating == 0 || sr.Date.IsZero() {
			continue
		}
		reviews = append(reviews, &entity.AppReview{
			ID:             uuid.New(),
			AppID:          app.ID,
			SourceReviewID: sr.SourceReviewID(),
			Author:         sr.Author,
			Rating:         sr.Rating,
			Body:           sr.Body,
			ReviewDate:     sr.Date,
			Location:       sr.Location,
			TimeUsing:      sr.TimeUsing,
			Source:         "shopify_app_store",
			ScrapedAt:      now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	if len(reviews) > 0 {
		if err := p.reviewRepo.UpsertBatch(ctx, reviews); err != nil {
			return fmt.Errorf("failed to store reviews: %w", err)
		}
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(reviews),
		Completed: len(reviews),
		Message:   fmt.Sprintf("Stored %d reviews", len(reviews)),
	})

	log.Printf("ReviewProcessor: stored %d reviews for app %s", len(reviews), payload.AppID)
	return p.syncJobRepo.MarkCompleted(ctx, payload.JobID)
}
