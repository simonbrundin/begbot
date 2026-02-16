package services

import (
	"sync"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
}

type FetchJob struct {
	ID               string
	Status           JobStatus
	Progress         int
	TotalQueries     int
	CompletedQueries int
	CurrentQuery     string
	AdsFound         int
	Error            string
	StartedAt        time.Time
	CompletedAt      time.Time
	Logs             []LogEntry
	logMu            sync.RWMutex
	logChannels      []chan LogEntry
	CancelChan       chan struct{}
}

type JobService struct {
	mu   sync.RWMutex
	jobs map[string]*FetchJob
}

func NewJobService() *JobService {
	return &JobService{
		jobs: make(map[string]*FetchJob),
	}
}

func (s *JobService) CreateJob(id string) *FetchJob {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &FetchJob{
		ID:         id,
		Status:     JobStatusPending,
		StartedAt:  time.Now(),
		CancelChan: make(chan struct{}),
	}
	s.jobs[id] = job
	return job
}

func (s *JobService) GetJob(id string) *FetchJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs[id]
}

func (s *JobService) StartJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = JobStatusRunning
	}
}

func (s *JobService) UpdateProgress(id string, completedQueries, totalQueries int, currentQuery string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.CompletedQueries = completedQueries
		job.TotalQueries = totalQueries
		job.CurrentQuery = currentQuery
		if totalQueries > 0 {
			job.Progress = (completedQueries * 100) / totalQueries
		}
	}
}

func (s *JobService) CompleteJob(id string, adsFound int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = JobStatusCompleted
		job.Progress = 100
		job.AdsFound = adsFound
		job.CompletedAt = time.Now()
	}
}

func (s *JobService) FailJob(id string, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = JobStatusFailed
		job.Error = errMsg
		job.CompletedAt = time.Now()
	}
}

func (s *JobService) CancelJob(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return false
	}

	// Can only cancel pending or running jobs
	if job.Status != JobStatusPending && job.Status != JobStatusRunning {
		return false
	}

	// Signal cancellation
	close(job.CancelChan)

	job.Status = JobStatusCancelled
	job.CompletedAt = time.Now()
	return true
}

func (s *JobService) AddLog(id string, level LogLevel, message string) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		return
	}

	job.logMu.Lock()
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}
	job.Logs = append(job.Logs, entry)

	if len(job.Logs) > 1000 {
		job.Logs = job.Logs[len(job.Logs)-1000:]
	}

	channels := make([]chan LogEntry, len(job.logChannels))
	copy(channels, job.logChannels)
	job.logMu.Unlock()

	for _, ch := range channels {
		select {
		case ch <- entry:
		default:
		}
	}
}

func (s *JobService) GetLogs(id string) []LogEntry {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	job.logMu.RLock()
	defer job.logMu.RUnlock()

	logs := make([]LogEntry, len(job.Logs))
	copy(logs, job.Logs)
	return logs
}

func (s *JobService) SubscribeToLogs(id string) chan LogEntry {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	ch := make(chan LogEntry, 100)

	job.logMu.Lock()
	job.logChannels = append(job.logChannels, ch)
	job.logMu.Unlock()

	return ch
}

func (s *JobService) UnsubscribeFromLogs(id string, ch chan LogEntry) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		return
	}

	job.logMu.Lock()
	for i, c := range job.logChannels {
		if c == ch {
			job.logChannels = append(job.logChannels[:i], job.logChannels[i+1:]...)
			close(c)
			break
		}
	}
	job.logMu.Unlock()
}
