package services

import (
	"context"
	"fmt"
	"time"

	"begbot/internal/models"
)

type CronJobRepository interface {
	CreateCronJob(ctx context.Context, job *models.CronJob) error
	GetAllCronJobs(ctx context.Context) ([]models.CronJob, error)
	GetCronJobByID(ctx context.Context, id int64) (*models.CronJob, error)
	UpdateCronJob(ctx context.Context, job *models.CronJob) error
	DeleteCronJob(ctx context.Context, id int64) error
}

type CronJobService struct {
	db CronJobRepository
}

func NewCronJobService(db CronJobRepository) *CronJobService {
	return &CronJobService{db: db}
}

func (s *CronJobService) CreateCronJob(ctx context.Context, job *models.CronJob) error {
	if err := validateCronExpression(job.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	return s.db.CreateCronJob(ctx, job)
}

func (s *CronJobService) GetAllCronJobs(ctx context.Context) ([]models.CronJob, error) {
	return s.db.GetAllCronJobs(ctx)
}

func (s *CronJobService) GetCronJobByID(ctx context.Context, id int64) (*models.CronJob, error) {
	return s.db.GetCronJobByID(ctx, id)
}

func (s *CronJobService) UpdateCronJob(ctx context.Context, job *models.CronJob) error {
	if err := validateCronExpression(job.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	job.UpdatedAt = time.Now()

	return s.db.UpdateCronJob(ctx, job)
}

func (s *CronJobService) DeleteCronJob(ctx context.Context, id int64) error {
	return s.db.DeleteCronJob(ctx, id)
}

func validateCronExpression(expr string) error {
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	fields := 0
	current := ""

	for _, ch := range expr {
		if ch == ' ' || ch == '\t' {
			if current != "" {
				fields++
				current = ""
			}
			continue
		}

		current += string(ch)

		if ch == ',' || ch == '-' || ch == '/' {
			current = ""
		}
	}

	if current != "" {
		fields++
	}

	if fields != 5 {
		return fmt.Errorf("expected 5 fields, got %d", fields)
	}

	return nil
}
