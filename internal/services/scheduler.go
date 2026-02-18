package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	db         *db.Postgres
	cron       *cron.Cron
	cfg        *config.Config
	botService *BotService
	running    map[int64]bool
}

func NewScheduler(db *db.Postgres, cfg *config.Config, botService *BotService) *Scheduler {
	return &Scheduler{
		db:         db,
		cron:       cron.New(),
		cfg:        cfg,
		botService: botService,
		running:    make(map[int64]bool),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	log.Println("Starting cron scheduler...")

	jobs, err := s.db.GetActiveCronJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cron jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.scheduleJob(job); err != nil {
			log.Printf("Failed to schedule job %d: %v", job.ID, err)
		}
	}

	s.cron.Start()
	log.Printf("Scheduler started with %d jobs", len(jobs))
	return nil
}

func (s *Scheduler) Stop() {
	log.Println("Stopping cron scheduler...")
	s.cron.Stop()
}

func (s *Scheduler) scheduleJob(job models.CronJob) error {
	_, err := s.cron.AddFunc(job.CronExpression, func() {
		s.runJob(job)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	log.Printf("Scheduled job %d: %s (%s)", job.ID, job.Name, job.CronExpression)
	return nil
}

func (s *Scheduler) runJob(job models.CronJob) {
	if s.running[job.ID] {
		log.Printf("Job %d (%s) is already running, skipping", job.ID, job.Name)
		return
	}

	s.running[job.ID] = true
	defer func() {
		s.running[job.ID] = false
	}()

	log.Printf("Starting job %d (%s)", job.ID, job.Name)

	searchTerms, err := s.db.GetActiveSearchTerms(context.Background())
	if err != nil {
		log.Printf("Error getting search terms for job %d: %v", job.ID, err)
		return
	}

	var filteredTerms []models.SearchTerm
	if len(job.SearchTermIDs) > 0 {
		termMap := make(map[int64]models.SearchTerm)
		for _, t := range searchTerms {
			termMap[t.ID] = t
		}
		for _, id := range job.SearchTermIDs {
			if t, ok := termMap[id]; ok {
				filteredTerms = append(filteredTerms, t)
			}
		}
	} else {
		filteredTerms = searchTerms
	}

	log.Printf("Job %d (%s): running with %d search terms", job.ID, job.Name, len(filteredTerms))

	if s.botService != nil {
		oldTerms := searchTerms
		s.botService.SetSearchTermsOverride(filteredTerms)

		if err := s.botService.Run(); err != nil {
			log.Printf("Error running job %d: %v", job.ID, err)
		}

		s.botService.SetSearchTermsOverride(oldTerms)
	}

	log.Printf("Completed job %d (%s)", job.ID, job.Name)
}

func (s *Scheduler) GetRunningJobs() map[int64]bool {
	return s.running
}

func (s *Scheduler) RefreshJobs(ctx context.Context) error {
	log.Println("Refreshing cron jobs...")

	jobs, err := s.db.GetActiveCronJobs(ctx)
	if err != nil {
		return err
	}

	s.cron.Stop()
	s.cron = cron.New()

	for _, job := range jobs {
		if err := s.scheduleJob(job); err != nil {
			log.Printf("Failed to schedule job %d: %v", job.ID, err)
		}
	}

	s.cron.Start()
	log.Printf("Refreshed %d jobs", len(jobs))
	return nil
}
