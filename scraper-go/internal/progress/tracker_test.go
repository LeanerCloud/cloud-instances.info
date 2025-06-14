package progress

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Status constants for testing
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

func TestTracker(t *testing.T) {
	// Create a temporary directory for tests
	tempDir := t.TempDir()
	progressFile := filepath.Join(tempDir, "progress.json")

	t.Run("NewTracker", func(t *testing.T) {
		tracker, err := NewTracker(progressFile)
		if err != nil {
			t.Fatalf("Failed to create tracker: %v", err)
		}
		if tracker == nil {
			t.Fatal("Tracker should not be nil")
		}
	})

	t.Run("StartService", func(t *testing.T) {
		tracker, _ := NewTracker(progressFile)
		
		err := tracker.StartService("ec2")
		if err != nil {
			t.Fatalf("Failed to start service: %v", err)
		}
		
		// Check if service was started  
		if !tracker.progress["ec2"].StartTime.After(time.Time{}) {
			t.Error("Start time should be set")
		}
		if tracker.progress["ec2"].Status != StatusInProgress {
			t.Errorf("Expected status %s, got %s", StatusInProgress, tracker.progress["ec2"].Status)
		}
	})

	t.Run("CompleteService", func(t *testing.T) {
		tracker, _ := NewTracker(progressFile)
		
		// Start then complete
		tracker.StartService("rds")
		err := tracker.CompleteService("rds", 50)
		if err != nil {
			t.Fatalf("Failed to complete service: %v", err)
		}
		
		service := tracker.progress["rds"]
		if service.Status != StatusCompleted {
			t.Errorf("Expected status %s, got %s", StatusCompleted, service.Status)
		}
		if service.InstancesFound != 50 {
			t.Errorf("Expected 50 instances, got %d", service.InstancesFound)
		}
		if service.EndTime == nil || service.EndTime.IsZero() {
			t.Error("End time should be set")
		}
	})

	t.Run("FailService", func(t *testing.T) {
		tracker, _ := NewTracker(progressFile)
		
		tracker.StartService("elasticache")
		err := tracker.FailService("elasticache", &TestError{msg: "test error"})
		if err != nil {
			t.Fatalf("Failed to fail service: %v", err)
		}
		
		service := tracker.progress["elasticache"]
		if service.Status != StatusFailed {
			t.Errorf("Expected status %s, got %s", StatusFailed, service.Status)
		}
		if service.LastError != "test error" {
			t.Errorf("Expected error 'test error', got '%s'", service.LastError)
		}
	})

	t.Run("IsServiceComplete", func(t *testing.T) {
		tracker, _ := NewTracker(progressFile)
		
		// Service not started should return false
		if tracker.IsServiceComplete("redshift") {
			t.Error("Non-started service should not be complete")
		}
		
		// Complete a service
		tracker.StartService("redshift")
		tracker.CompleteService("redshift", 10)
		
		if !tracker.IsServiceComplete("redshift") {
			t.Error("Completed service should return true")
		}
	})

	t.Run("Summary", func(t *testing.T) {
		// Use a separate file to avoid test interference
		summaryProgressFile := filepath.Join(tempDir, "summary_progress.json")
		tracker, _ := NewTracker(summaryProgressFile)
		
		// Add some services in different states
		tracker.StartService("ec2")
		tracker.CompleteService("ec2", 100)
		
		tracker.StartService("rds")
		tracker.FailService("rds", &TestError{msg: "failed"})
		
		tracker.StartService("elasticache")
		// Leave in progress
		
		summary := tracker.Summary()
		if summary == "" {
			t.Error("Summary should not be empty")
		}
		
		// Check that it contains expected information
		expectedParts := []string{"1 completed", "1 in progress", "1 failed"}
		for _, part := range expectedParts {
			if !strings.Contains(summary, part) {
				t.Errorf("Summary should contain '%s': %s", part, summary)
			}
		}
	})

	t.Run("PersistenceAndLoad", func(t *testing.T) {
		// Create tracker and add some progress
		tracker1, _ := NewTracker(progressFile)
		tracker1.StartService("opensearch")
		tracker1.CompleteService("opensearch", 25)
		
		// Create new tracker from same file
		tracker2, err := NewTracker(progressFile)
		if err != nil {
			t.Fatalf("Failed to load tracker: %v", err)
		}
		
		// Check that progress was loaded
		if !tracker2.IsServiceComplete("opensearch") {
			t.Error("Progress should have been loaded from file")
		}
		
		service := tracker2.progress["opensearch"]
		if service.InstancesFound != 25 {
			t.Errorf("Expected 25 instances, got %d", service.InstancesFound)
		}
	})

	t.Run("InvalidFile", func(t *testing.T) {
		// Try to create tracker with invalid directory (should succeed, but fail on save)
		invalidPath := filepath.Join("/root/nonexistent/progress.json")
		tracker, err := NewTracker(invalidPath)
		if err != nil {
			t.Fatalf("NewTracker should not fail: %v", err)
		}
		
		// Should fail when trying to save
		err = tracker.StartService("test")
		if err == nil {
			t.Error("Should fail when trying to save to invalid path")
		}
	})

	t.Run("GetProgress", func(t *testing.T) {
		getProgressFile := filepath.Join(tempDir, "get_progress.json")
		tracker, _ := NewTracker(getProgressFile)
		
		// Start some services
		tracker.StartService("ec2")
		tracker.CompleteService("ec2", 50)
		tracker.StartService("rds")
		
		// Get progress
		progress := tracker.GetProgress()
		if len(progress) != 2 {
			t.Errorf("Expected 2 services, got %d", len(progress))
		}
		
		// Check that we got copies, not originals
		if progress["ec2"].InstancesFound != 50 {
			t.Errorf("Expected 50 instances for ec2, got %d", progress["ec2"].InstancesFound)
		}
	})

	t.Run("UpdateService", func(t *testing.T) {
		updateServiceFile := filepath.Join(tempDir, "update_service.json")
		tracker, _ := NewTracker(updateServiceFile)
		
		// Start a service
		tracker.StartService("opensearch")
		
		// Update it
		err := tracker.UpdateService("opensearch", func(sp *ServiceProgress) {
			if sp.Metadata == nil {
				sp.Metadata = make(map[string]interface{})
			}
			sp.Metadata["regions"] = []string{"us-east-1", "eu-west-1"}
			sp.Metadata["pricing"] = true
		})
		if err != nil {
			t.Fatalf("Failed to update service: %v", err)
		}
		
		// Check that metadata was set
		progress := tracker.GetProgress()
		service := progress["opensearch"]
		if service.Metadata["pricing"] != true {
			t.Error("Metadata should have been updated")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		resetFile := filepath.Join(tempDir, "reset.json")
		tracker, _ := NewTracker(resetFile)
		
		// Add some progress
		tracker.StartService("ec2")
		tracker.CompleteService("ec2", 100)
		
		// Reset
		err := tracker.Reset()
		if err != nil {
			t.Fatalf("Failed to reset tracker: %v", err)
		}
		
		// Check that progress is empty
		progress := tracker.GetProgress()
		if len(progress) != 0 {
			t.Errorf("Expected 0 services after reset, got %d", len(progress))
		}
	})
}

// Helper types and functions for testing
type TestError struct {
	msg string
}

func (e *TestError) Error() string {
	return e.msg
}

