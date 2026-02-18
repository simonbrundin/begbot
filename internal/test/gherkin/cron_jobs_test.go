package gherkin

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

type cronJobTestState struct {
	service     *services.CronJobService
	mockDB      *mockCronJobDB
	resultJobs  []models.CronJob
	resultJob   *models.CronJob
	resultErr   error
	resultCount int
}

type mockCronJobDB struct {
	jobs []models.CronJob
	err  error
}

func (m *mockCronJobDB) CreateCronJob(ctx context.Context, job *models.CronJob) error {
	if m.err != nil {
		return m.err
	}
	job.ID = int64(len(m.jobs) + 1)
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	m.jobs = append(m.jobs, *job)
	return nil
}

func (m *mockCronJobDB) GetAllCronJobs(ctx context.Context) ([]models.CronJob, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.jobs, nil
}

func (m *mockCronJobDB) GetCronJobByID(ctx context.Context, id int64) (*models.CronJob, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, job := range m.jobs {
		if job.ID == id {
			return &job, nil
		}
	}
	return nil, nil
}

func (m *mockCronJobDB) UpdateCronJob(ctx context.Context, job *models.CronJob) error {
	if m.err != nil {
		return m.err
	}
	for i, j := range m.jobs {
		if j.ID == job.ID {
			job.UpdatedAt = time.Now()
			m.jobs[i] = *job
			return nil
		}
	}
	return errors.New("cron job not found")
}

func (m *mockCronJobDB) DeleteCronJob(ctx context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	for i, job := range m.jobs {
		if job.ID == id {
			m.jobs = append(m.jobs[:i], m.jobs[i+1:]...)
			return nil
		}
	}
	return errors.New("cron job not found")
}

func InitializeCronJobScenario(ctx *godog.ScenarioContext) {
	state := &cronJobTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &cronJobTestState{
			mockDB: &mockCronJobDB{jobs: []models.CronJob{}},
		}
		state.service = services.NewCronJobService(state.mockDB)
	})

	ctx.Given("a cron job service with mock database", func() error {
		state.mockDB = &mockCronJobDB{jobs: []models.CronJob{}}
		state.service = services.NewCronJobService(state.mockDB)
		return nil
	})

	ctx.Given("the database has cron job records", func(table *godog.Table) error {
		for _, row := range table.Rows {
			id := int64(0)
			fmt.Sscanf(row.Cells[0].Value, "%d", &id)

			// Parse search term IDs from format like "[1,2]" or "[]"
			var searchTermIDs []int64
			idsStr := row.Cells[3].Value
			if idsStr == "[]" || idsStr == "" {
				searchTermIDs = []int64{}
			} else {
				re := regexp.MustCompile(`\d+`)
				matches := re.FindAllString(idsStr, -1)
				for _, match := range matches {
					var termID int64
					fmt.Sscanf(match, "%d", &termID)
					searchTermIDs = append(searchTermIDs, termID)
				}
			}

			isActive := row.Cells[4].Value == "true"

			state.mockDB.jobs = append(state.mockDB.jobs, models.CronJob{
				ID:             id,
				Name:           row.Cells[1].Value,
				CronExpression: row.Cells[2].Value,
				SearchTermIDs:  searchTermIDs,
				IsActive:       isActive,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}
		return nil
	})

	ctx.When("I create a cron job with name {string}, expression {string}, search term IDs {string}, and active {bool}", func(name, expr, termIDs string, active bool) error {
		ctx := context.Background()

		var searchTermIDs []int64
		if termIDs != "[]" && termIDs != "" {
			re := regexp.MustCompile(`\d+`)
			matches := re.FindAllString(termIDs, -1)
			for _, match := range matches {
				var id int64
				fmt.Sscanf(match, "%d", &id)
				searchTermIDs = append(searchTermIDs, id)
			}
		}

		job := &models.CronJob{
			Name:           name,
			CronExpression: expr,
			SearchTermIDs:  searchTermIDs,
			IsActive:       active,
		}

		state.resultErr = state.service.CreateCronJob(ctx, job)
		if state.resultErr == nil {
			state.resultJob = job
		}
		return nil
	})

	ctx.When("I get all cron jobs", func() error {
		ctx := context.Background()
		state.resultJobs, state.resultErr = state.service.GetAllCronJobs(ctx)
		state.resultCount = len(state.resultJobs)
		return nil
	})

	ctx.When("I update cron job {int} to name {string}, expression {string}, search term IDs {string}, active {bool}", func(id int64, name, expr, termIDs string, active bool) error {
		ctx := context.Background()

		var searchTermIDs []int64
		if termIDs != "[]" && termIDs != "" {
			re := regexp.MustCompile(`\d+`)
			matches := re.FindAllString(termIDs, -1)
			for _, match := range matches {
				var tid int64
				fmt.Sscanf(match, "%d", &tid)
				searchTermIDs = append(searchTermIDs, tid)
			}
		}

		job := &models.CronJob{
			ID:             id,
			Name:           name,
			CronExpression: expr,
			SearchTermIDs:  searchTermIDs,
			IsActive:       active,
		}
		state.resultErr = state.service.UpdateCronJob(ctx, job)
		if state.resultErr == nil {
			state.resultJob = job
		}
		return nil
	})

	ctx.When("I delete cron job {int}", func(id int64) error {
		ctx := context.Background()
		state.resultErr = state.service.DeleteCronJob(ctx, id)
		return nil
	})

	ctx.When("I toggle cron job {int} active status", func(id int64) error {
		ctx := context.Background()
		job, err := state.service.GetCronJobByID(ctx, id)
		if err != nil {
			state.resultErr = err
			return nil
		}
		job.IsActive = !job.IsActive
		state.resultErr = state.service.UpdateCronJob(ctx, job)
		state.resultJob = job
		return nil
	})

	ctx.When("I get cron job by ID {int}", func(id int64) error {
		ctx := context.Background()
		job, err := state.service.GetCronJobByID(ctx, id)
		state.resultErr = err
		if job != nil {
			state.resultJobs = []models.CronJob{*job}
			state.resultCount = 1
		} else {
			state.resultJobs = []models.CronJob{}
			state.resultCount = 0
		}
		return nil
	})

	ctx.Then("the cron job should be saved successfully", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})

	ctx.Then("the cron job should have ID set", func() error {
		if state.resultJob == nil || state.resultJob.ID == 0 {
			return fmt.Errorf("expected ID to be set")
		}
		return nil
	})

	ctx.Then("the cron job name should be {string}", func(expected string) error {
		if state.resultJob == nil || state.resultJob.Name != expected {
			return fmt.Errorf("expected name '%s', got '%s'", expected, state.resultJob.Name)
		}
		return nil
	})

	ctx.Then("the cron job expression should be {string}", func(expected string) error {
		if state.resultJob == nil || state.resultJob.CronExpression != expected {
			return fmt.Errorf("expected expression '%s', got '%s'", expected, state.resultJob.CronExpression)
		}
		return nil
	})

	ctx.Then("the cron job should be active", func() error {
		if state.resultJob == nil || !state.resultJob.IsActive {
			return fmt.Errorf("expected IsActive to be true")
		}
		return nil
	})

	ctx.Then("the cron job should be inactive", func() error {
		if state.resultJob == nil || state.resultJob.IsActive {
			return fmt.Errorf("expected IsActive to be false")
		}
		return nil
	})

	ctx.Then("I should receive {int} cron job records", func(count int) error {
		if len(state.resultJobs) != count {
			return fmt.Errorf("expected %d records, got %d", count, len(state.resultJobs))
		}
		return nil
	})

	ctx.Then("the first cron job should have name {string}", func(expected string) error {
		if len(state.resultJobs) == 0 || state.resultJobs[0].Name != expected {
			return fmt.Errorf("expected first job name '%s', got '%s'", expected, state.resultJobs[0].Name)
		}
		return nil
	})

	ctx.Then("the cron job should be updated successfully", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})

	ctx.Then("the cron job should be deleted successfully", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})

	ctx.Then("there should be {int} cron jobs in the database", func(count int) error {
		if len(state.mockDB.jobs) != count {
			return fmt.Errorf("expected %d jobs in DB, got %d", count, len(state.mockDB.jobs))
		}
		return nil
	})

	ctx.Then("the cron job should have empty search term IDs", func() error {
		if state.resultJob == nil || len(state.resultJob.SearchTermIDs) != 0 {
			return fmt.Errorf("expected empty SearchTermIDs, got %v", state.resultJob.SearchTermIDs)
		}
		return nil
	})

	ctx.Then("an error should be returned", func() error {
		if state.resultErr == nil {
			return fmt.Errorf("expected error, got nil")
		}
		return nil
	})

	ctx.Then("the error message should contain {string}", func(expected string) error {
		if state.resultErr == nil || !strings.Contains(state.resultErr.Error(), expected) {
			return fmt.Errorf("expected error containing '%s', got '%s'", expected, state.resultErr.Error())
		}
		return nil
	})
}

func getCronJobFeaturesPath(filename string) string {
	cwd, _ := getCwd()
	return cwd + "/features/" + filename
}

func getCwd() (string, error) {
	return "/home/simon/repos/begbot/internal/test/gherkin", nil
}

func TestCronJobFeatures(t *testing.T) {
	featurePath := getCronJobFeaturesPath("cron_jobs.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeCronJobScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run cron jobs gherkin tests")
	}
}
