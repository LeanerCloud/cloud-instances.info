package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Tracker tracks the progress of scraping operations
type Tracker struct {
	mu       sync.RWMutex
	file     string
	progress map[string]*ServiceProgress
}

// ServiceProgress tracks progress for a specific service
type ServiceProgress struct {
	Service          string                 `json:"service"`
	Status           string                 `json:"status"` // pending, in_progress, completed, failed
	StartTime        *time.Time             `json:"start_time,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	InstancesFound   int                    `json:"instances_found"`
	PricingComplete  bool                   `json:"pricing_complete"`
	SpotComplete     bool                   `json:"spot_complete"`
	LastError        string                 `json:"last_error,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewTracker creates a new progress tracker
func NewTracker(progressFile string) (*Tracker, error) {
	t := &Tracker{
		file:     progressFile,
		progress: make(map[string]*ServiceProgress),
	}
	
	// Load existing progress if file exists
	if err := t.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	
	return t, nil
}

// load reads progress from disk
func (t *Tracker) load() error {
	data, err := os.ReadFile(t.file)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, &t.progress)
}

// save writes progress to disk
func (t *Tracker) save() error {
	// Ensure directory exists
	dir := filepath.Dir(t.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(t.progress, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(t.file, data, 0644)
}

// StartService marks a service as started
func (t *Tracker) StartService(service string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	now := time.Now()
	t.progress[service] = &ServiceProgress{
		Service:   service,
		Status:    "in_progress",
		StartTime: &now,
		Metadata:  make(map[string]interface{}),
	}
	
	return t.save()
}

// UpdateService updates progress for a service
func (t *Tracker) UpdateService(service string, updates func(*ServiceProgress)) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	sp, ok := t.progress[service]
	if !ok {
		sp = &ServiceProgress{
			Service:  service,
			Status:   "in_progress",
			Metadata: make(map[string]interface{}),
		}
		t.progress[service] = sp
	}
	
	updates(sp)
	
	return t.save()
}

// CompleteService marks a service as completed
func (t *Tracker) CompleteService(service string, instancesFound int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	sp, ok := t.progress[service]
	if !ok {
		sp = &ServiceProgress{
			Service:  service,
			Metadata: make(map[string]interface{}),
		}
		t.progress[service] = sp
	}
	
	now := time.Now()
	sp.Status = "completed"
	sp.EndTime = &now
	sp.InstancesFound = instancesFound
	sp.LastError = ""
	
	return t.save()
}

// FailService marks a service as failed
func (t *Tracker) FailService(service string, err error) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	sp, ok := t.progress[service]
	if !ok {
		sp = &ServiceProgress{
			Service:  service,
			Metadata: make(map[string]interface{}),
		}
		t.progress[service] = sp
	}
	
	now := time.Now()
	sp.Status = "failed"
	sp.EndTime = &now
	sp.LastError = err.Error()
	
	return t.save()
}

// GetProgress returns the current progress
func (t *Tracker) GetProgress() map[string]*ServiceProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// Return a copy to avoid concurrent modification
	result := make(map[string]*ServiceProgress)
	for k, v := range t.progress {
		result[k] = v
	}
	
	return result
}

// IsServiceComplete checks if a service has been completed
func (t *Tracker) IsServiceComplete(service string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	sp, ok := t.progress[service]
	if !ok {
		return false
	}
	
	return sp.Status == "completed"
}

// Reset clears all progress
func (t *Tracker) Reset() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.progress = make(map[string]*ServiceProgress)
	return t.save()
}

// Summary returns a summary of the current progress
func (t *Tracker) Summary() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	completed := 0
	failed := 0
	inProgress := 0
	pending := 0
	
	for _, sp := range t.progress {
		switch sp.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "in_progress":
			inProgress++
		case "pending":
			pending++
		}
	}
	
	return fmt.Sprintf("Progress: %d completed, %d in progress, %d failed, %d pending",
		completed, inProgress, failed, pending)
}