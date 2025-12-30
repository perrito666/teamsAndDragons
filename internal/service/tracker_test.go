package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"teamsAndDragons/internal/domain"
	"teamsAndDragons/internal/storage"
)

func setupTestTracker(t *testing.T) (*Tracker, string) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "tracker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create storage
	store, err := storage.NewFileStorage(tempDir, nil)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create exporter
	exporter := storage.NewHTMLExporter(store)

	// Create tracker
	tracker := NewTracker(store, exporter, nil)

	return tracker, tempDir
}

func TestTracker_CreatePerson(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, err := tracker.CreatePerson("John Doe", "Test notes")
	if err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	if person.ID == "" {
		t.Error("CreatePerson() ID should not be empty")
	}
	if person.Name != "John Doe" {
		t.Errorf("CreatePerson() Name = %v, want John Doe", person.Name)
	}
	if person.Notes != "Test notes" {
		t.Errorf("CreatePerson() Notes = %v, want 'Test notes'", person.Notes)
	}

	// Verify file was created
	profilePath := filepath.Join(tempDir, "john-doe", "00-profile.md")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Error("Profile file was not created")
	}
}

func TestTracker_ListPeople(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	// Create some people
	_, err := tracker.CreatePerson("Alice", "")
	if err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}
	_, err = tracker.CreatePerson("Bob", "")
	if err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	people, err := tracker.ListPeople()
	if err != nil {
		t.Fatalf("ListPeople() error = %v", err)
	}

	if len(people) != 2 {
		t.Errorf("ListPeople() count = %v, want 2", len(people))
	}

	// Should be sorted by name
	if people[0].Name != "Alice" {
		t.Errorf("ListPeople()[0].Name = %v, want Alice", people[0].Name)
	}
	if people[1].Name != "Bob" {
		t.Errorf("ListPeople()[1].Name = %v, want Bob", people[1].Name)
	}
}

func TestTracker_UpdatePerson(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, err := tracker.CreatePerson("John", "Original notes")
	if err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	updated, err := tracker.UpdatePerson(person.ID, "John Doe", "Updated notes")
	if err != nil {
		t.Fatalf("UpdatePerson() error = %v", err)
	}

	if updated.Name != "John Doe" {
		t.Errorf("UpdatePerson() Name = %v, want John Doe", updated.Name)
	}
	if updated.Notes != "Updated notes" {
		t.Errorf("UpdatePerson() Notes = %v, want 'Updated notes'", updated.Notes)
	}
}

func TestTracker_DeletePerson(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, err := tracker.CreatePerson("John", "")
	if err != nil {
		t.Fatalf("CreatePerson() error = %v", err)
	}

	err = tracker.DeletePerson(person.ID)
	if err != nil {
		t.Fatalf("DeletePerson() error = %v", err)
	}

	// Verify person is gone
	_, err = tracker.GetPerson(person.ID)
	if err != domain.ErrPersonNotFound {
		t.Errorf("GetPerson() after delete error = %v, want ErrPersonNotFound", err)
	}
}

func TestTracker_CreateMilestone(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")

	milestone, err := tracker.CreateMilestone(person.ID, "Senior 1 → Senior 2", "Test description")
	if err != nil {
		t.Fatalf("CreateMilestone() error = %v", err)
	}

	if milestone.ID == "" {
		t.Error("CreateMilestone() ID should not be empty")
	}
	if milestone.Name != "Senior 1 → Senior 2" {
		t.Errorf("CreateMilestone() Name = %v, want 'Senior 1 → Senior 2'", milestone.Name)
	}
	if milestone.Number != 1 {
		t.Errorf("CreateMilestone() Number = %v, want 1", milestone.Number)
	}
	if milestone.Status != domain.MilestoneInProgress {
		t.Errorf("CreateMilestone() Status = %v, want in_progress", milestone.Status)
	}
}

func TestTracker_MilestoneNumbering(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")

	m1, _ := tracker.CreateMilestone(person.ID, "First", "")
	m2, _ := tracker.CreateMilestone(person.ID, "Second", "")
	m3, _ := tracker.CreateMilestone(person.ID, "Third", "")

	if m1.Number != 1 {
		t.Errorf("First milestone Number = %v, want 1", m1.Number)
	}
	if m2.Number != 2 {
		t.Errorf("Second milestone Number = %v, want 2", m2.Number)
	}
	if m3.Number != 3 {
		t.Errorf("Third milestone Number = %v, want 3", m3.Number)
	}
}

func TestTracker_AddObjective(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")

	updated, err := tracker.AddObjective(milestone.ID, "Better communication", "Test description")
	if err != nil {
		t.Fatalf("AddObjective() error = %v", err)
	}

	if len(updated.Objectives) != 1 {
		t.Fatalf("AddObjective() Objectives count = %v, want 1", len(updated.Objectives))
	}

	obj := updated.Objectives[0]
	if obj.Name != "Better communication" {
		t.Errorf("AddObjective() Name = %v, want 'Better communication'", obj.Name)
	}
	if obj.Achieved {
		t.Error("AddObjective() new objective should not be achieved")
	}
}

func TestTracker_AddAction(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Test Objective", "")

	objID := milestone.Objectives[0].ID
	actionDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	updated, err := tracker.AddAction(milestone.ID, objID, "Led workshop", domain.ImpactPositive, actionDate, "Notes")
	if err != nil {
		t.Fatalf("AddAction() error = %v", err)
	}

	if len(updated.Objectives[0].Actions) != 1 {
		t.Fatalf("AddAction() Actions count = %v, want 1", len(updated.Objectives[0].Actions))
	}

	act := updated.Objectives[0].Actions[0]
	if act.Description != "Led workshop" {
		t.Errorf("AddAction() Description = %v, want 'Led workshop'", act.Description)
	}
	if act.Impact != domain.ImpactPositive {
		t.Errorf("AddAction() Impact = %v, want positive", act.Impact)
	}
}

func TestTracker_SetObjectiveAchieved(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Test Objective", "")

	objID := milestone.Objectives[0].ID

	// Mark as achieved
	updated, err := tracker.SetObjectiveAchieved(milestone.ID, objID, true, "Great work!")
	if err != nil {
		t.Fatalf("SetObjectiveAchieved() error = %v", err)
	}

	obj := updated.Objectives[0]
	if !obj.Achieved {
		t.Error("SetObjectiveAchieved() objective should be achieved")
	}
	if obj.ReviewNotes != "Great work!" {
		t.Errorf("SetObjectiveAchieved() ReviewNotes = %v, want 'Great work!'", obj.ReviewNotes)
	}

	// Mark as not achieved
	updated, err = tracker.SetObjectiveAchieved(milestone.ID, objID, false, "")
	if err != nil {
		t.Fatalf("SetObjectiveAchieved() error = %v", err)
	}

	if updated.Objectives[0].Achieved {
		t.Error("SetObjectiveAchieved() objective should not be achieved")
	}
}

func TestTracker_GetObjectiveReview(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Test Objective", "")

	objID := milestone.Objectives[0].ID
	now := time.Now()

	// Add some actions
	tracker.AddAction(milestone.ID, objID, "Positive 1", domain.ImpactPositive, now, "")
	tracker.AddAction(milestone.ID, objID, "Positive 2", domain.ImpactPositive, now, "")
	tracker.AddAction(milestone.ID, objID, "Negative 1", domain.ImpactNegative, now, "")

	review, err := tracker.GetObjectiveReview(milestone.ID, objID)
	if err != nil {
		t.Fatalf("GetObjectiveReview() error = %v", err)
	}

	if review.PositiveCount != 2 {
		t.Errorf("GetObjectiveReview() PositiveCount = %v, want 2", review.PositiveCount)
	}
	if review.NegativeCount != 1 {
		t.Errorf("GetObjectiveReview() NegativeCount = %v, want 1", review.NegativeCount)
	}
	if review.Net != 1 {
		t.Errorf("GetObjectiveReview() Net = %v, want 1", review.Net)
	}
}

func TestTracker_DeleteObjective(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Objective 1", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Objective 2", "")

	objID := milestone.Objectives[0].ID

	updated, err := tracker.DeleteObjective(milestone.ID, objID)
	if err != nil {
		t.Fatalf("DeleteObjective() error = %v", err)
	}

	if len(updated.Objectives) != 1 {
		t.Errorf("DeleteObjective() Objectives count = %v, want 1", len(updated.Objectives))
	}
	if updated.Objectives[0].Name != "Objective 2" {
		t.Errorf("DeleteObjective() remaining objective Name = %v, want 'Objective 2'", updated.Objectives[0].Name)
	}
}

func TestTracker_DeleteAction(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John", "")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "")
	milestone, _ = tracker.AddObjective(milestone.ID, "Test Objective", "")

	objID := milestone.Objectives[0].ID
	now := time.Now()

	milestone, _ = tracker.AddAction(milestone.ID, objID, "Action 1", domain.ImpactPositive, now, "")
	milestone, _ = tracker.AddAction(milestone.ID, objID, "Action 2", domain.ImpactNegative, now, "")

	actID := milestone.Objectives[0].Actions[0].ID

	updated, err := tracker.DeleteAction(milestone.ID, objID, actID)
	if err != nil {
		t.Fatalf("DeleteAction() error = %v", err)
	}

	if len(updated.Objectives[0].Actions) != 1 {
		t.Errorf("DeleteAction() Actions count = %v, want 1", len(updated.Objectives[0].Actions))
	}
}

func TestTracker_ExportPerson(t *testing.T) {
	tracker, tempDir := setupTestTracker(t)
	defer os.RemoveAll(tempDir)

	person, _ := tracker.CreatePerson("John Doe", "Some notes")
	milestone, _ := tracker.CreateMilestone(person.ID, "Test Milestone", "Description")
	tracker.AddObjective(milestone.ID, "Test Objective", "Obj description")

	outputPath := filepath.Join(tempDir, "export.html")
	err := tracker.ExportPerson(person.ID, outputPath, true)
	if err != nil {
		t.Fatalf("ExportPerson() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Check content
	if !contains(string(content), "John Doe") {
		t.Error("Export missing person name")
	}
	if !contains(string(content), "Test Milestone") {
		t.Error("Export missing milestone")
	}
	if !contains(string(content), "Test Objective") {
		t.Error("Export missing objective")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
