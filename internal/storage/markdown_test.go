package storage

import (
	"strings"
	"testing"
	"time"

	"teamsAndDragons/internal/domain"
)

func TestParseProfile(t *testing.T) {
	content := `---
id: "abc123"
name: "John Doe"
created_at: 2024-01-15T10:00:00Z
updated_at: 2024-03-20T14:30:00Z
---

# John Doe

Some notes about John.
He is a great engineer.
`

	person, err := ParseProfile([]byte(content))
	if err != nil {
		t.Fatalf("ParseProfile() error = %v", err)
	}

	if person.ID != "abc123" {
		t.Errorf("ID = %v, want abc123", person.ID)
	}
	if person.Name != "John Doe" {
		t.Errorf("Name = %v, want John Doe", person.Name)
	}
	if !strings.Contains(person.Notes, "great engineer") {
		t.Errorf("Notes = %v, want to contain 'great engineer'", person.Notes)
	}
}

func TestParseProfile_InvalidFrontmatter(t *testing.T) {
	content := `No frontmatter here
Just some text`

	_, err := ParseProfile([]byte(content))
	if err == nil {
		t.Error("ParseProfile() expected error for invalid frontmatter")
	}
}

func TestRenderProfile(t *testing.T) {
	person := &domain.Person{
		ID:        "abc123",
		Name:      "John Doe",
		Notes:     "Some notes about John.",
		CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 3, 20, 14, 30, 0, 0, time.UTC),
	}

	content := RenderProfile(person)

	// Should contain frontmatter markers
	if !strings.Contains(string(content), "---") {
		t.Error("RenderProfile() missing frontmatter markers")
	}

	// Should contain ID
	if !strings.Contains(string(content), "id: abc123") {
		t.Error("RenderProfile() missing ID")
	}

	// Should contain name
	if !strings.Contains(string(content), "name: John Doe") {
		t.Error("RenderProfile() missing name")
	}

	// Should contain title
	if !strings.Contains(string(content), "# John Doe") {
		t.Error("RenderProfile() missing title")
	}

	// Should contain notes
	if !strings.Contains(string(content), "Some notes about John.") {
		t.Error("RenderProfile() missing notes")
	}
}

func TestParseMilestone(t *testing.T) {
	content := `---
id: "xyz789"
number: 1
name: "Senior 1 → Senior 2"
status: "in_progress"
created_at: 2024-01-15T10:00:00Z
updated_at: 2024-03-20T14:30:00Z
---

# Milestone: Senior 1 → Senior 2

Overall description of the milestone.

## Objectives

### Better communication
<!-- id: obj-001 -->
<!-- achieved: false -->

Demonstrate ability to share knowledge.

> **Review Notes:** Showing progress.

#### Actions

- **[+] 2024-02-15:** Led team workshop
  <!-- id: act-001 -->
  Well received by team.

- **[-] 2024-03-01:** Missed deadline
  <!-- id: act-002 -->
  Was overloaded.

### Technical leadership
<!-- id: obj-002 -->
<!-- achieved: true -->

Take ownership of technical decisions.

#### Actions

- **[+] 2024-01-20:** Led architecture review
  <!-- id: act-003 -->
  Great impact.
`

	milestone, err := ParseMilestone([]byte(content), "person-123")
	if err != nil {
		t.Fatalf("ParseMilestone() error = %v", err)
	}

	if milestone.ID != "xyz789" {
		t.Errorf("ID = %v, want xyz789", milestone.ID)
	}
	if milestone.Number != 1 {
		t.Errorf("Number = %v, want 1", milestone.Number)
	}
	if milestone.Name != "Senior 1 → Senior 2" {
		t.Errorf("Name = %v, want 'Senior 1 → Senior 2'", milestone.Name)
	}
	if milestone.Status != domain.MilestoneInProgress {
		t.Errorf("Status = %v, want in_progress", milestone.Status)
	}
	if milestone.PersonID != "person-123" {
		t.Errorf("PersonID = %v, want person-123", milestone.PersonID)
	}
	if !strings.Contains(milestone.Description, "Overall description") {
		t.Errorf("Description = %v, want to contain 'Overall description'", milestone.Description)
	}

	// Check objectives
	if len(milestone.Objectives) != 2 {
		t.Fatalf("Objectives count = %v, want 2", len(milestone.Objectives))
	}

	obj1 := milestone.Objectives[0]
	if obj1.Name != "Better communication" {
		t.Errorf("Objective[0].Name = %v, want 'Better communication'", obj1.Name)
	}
	if obj1.ID != "obj-001" {
		t.Errorf("Objective[0].ID = %v, want obj-001", obj1.ID)
	}
	if obj1.Achieved {
		t.Error("Objective[0].Achieved = true, want false")
	}
	if obj1.ReviewNotes != "Showing progress." {
		t.Errorf("Objective[0].ReviewNotes = %v, want 'Showing progress.'", obj1.ReviewNotes)
	}

	// Check actions
	if len(obj1.Actions) != 2 {
		t.Fatalf("Objective[0].Actions count = %v, want 2", len(obj1.Actions))
	}

	act1 := obj1.Actions[0]
	if act1.Impact != domain.ImpactPositive {
		t.Errorf("Action[0].Impact = %v, want positive", act1.Impact)
	}
	if act1.Description != "Led team workshop" {
		t.Errorf("Action[0].Description = %v, want 'Led team workshop'", act1.Description)
	}
	if act1.ID != "act-001" {
		t.Errorf("Action[0].ID = %v, want act-001", act1.ID)
	}

	act2 := obj1.Actions[1]
	if act2.Impact != domain.ImpactNegative {
		t.Errorf("Action[1].Impact = %v, want negative", act2.Impact)
	}

	obj2 := milestone.Objectives[1]
	if !obj2.Achieved {
		t.Error("Objective[1].Achieved = false, want true")
	}
}

func TestRenderMilestone(t *testing.T) {
	milestone := &domain.Milestone{
		ID:          "xyz789",
		Number:      1,
		Name:        "Senior 1 → Senior 2",
		Description: "Overall description.",
		Status:      domain.MilestoneInProgress,
		CreatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 3, 20, 14, 30, 0, 0, time.UTC),
		Objectives: []domain.Objective{
			{
				ID:          "obj-001",
				Name:        "Better communication",
				Description: "Demonstrate knowledge sharing.",
				Achieved:    false,
				ReviewNotes: "Making progress.",
				Actions: []domain.Action{
					{
						ID:          "act-001",
						Description: "Led workshop",
						Impact:      domain.ImpactPositive,
						Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
						Notes:       "Well received.",
					},
				},
			},
		},
	}

	content := RenderMilestone(milestone)
	contentStr := string(content)

	// Check frontmatter
	if !strings.Contains(contentStr, "id: xyz789") {
		t.Error("RenderMilestone() missing ID")
	}
	if !strings.Contains(contentStr, "number: 1") {
		t.Error("RenderMilestone() missing number")
	}
	if !strings.Contains(contentStr, "status: in_progress") {
		t.Error("RenderMilestone() missing status")
	}

	// Check title
	if !strings.Contains(contentStr, "# Milestone: Senior 1 → Senior 2") {
		t.Error("RenderMilestone() missing title")
	}

	// Check objective
	if !strings.Contains(contentStr, "### Better communication") {
		t.Error("RenderMilestone() missing objective heading")
	}
	if !strings.Contains(contentStr, "<!-- id: obj-001 -->") {
		t.Error("RenderMilestone() missing objective ID comment")
	}
	if !strings.Contains(contentStr, "<!-- achieved: false -->") {
		t.Error("RenderMilestone() missing achieved comment")
	}

	// Check action
	if !strings.Contains(contentStr, "**[+] 2024-02-15:**") {
		t.Error("RenderMilestone() missing action date")
	}
	if !strings.Contains(contentStr, "Led workshop") {
		t.Error("RenderMilestone() missing action description")
	}
}

func TestRoundTrip_Profile(t *testing.T) {
	original := &domain.Person{
		ID:        "test-123",
		Name:      "Test Person",
		Notes:     "Some test notes\nwith multiple lines.",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
	}

	// Render
	content := RenderProfile(original)

	// Parse back
	parsed, err := ParseProfile(content)
	if err != nil {
		t.Fatalf("Round trip parse error: %v", err)
	}

	// Compare
	if parsed.ID != original.ID {
		t.Errorf("Round trip ID: got %v, want %v", parsed.ID, original.ID)
	}
	if parsed.Name != original.Name {
		t.Errorf("Round trip Name: got %v, want %v", parsed.Name, original.Name)
	}
	if !strings.Contains(parsed.Notes, "test notes") {
		t.Errorf("Round trip Notes: got %v, want to contain 'test notes'", parsed.Notes)
	}
}

func TestRoundTrip_Milestone(t *testing.T) {
	original := &domain.Milestone{
		ID:          "ms-123",
		PersonID:    "person-123",
		Number:      2,
		Name:        "Test Milestone",
		Description: "Test description",
		Status:      domain.MilestoneCompleted,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Objectives: []domain.Objective{
			{
				ID:          "obj-1",
				Name:        "Test Objective",
				Description: "Objective description",
				Achieved:    true,
				ReviewNotes: "Great work!",
				Actions: []domain.Action{
					{
						ID:          "act-1",
						Description: "Test action",
						Impact:      domain.ImpactPositive,
						Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
						Notes:       "Action notes",
					},
					{
						ID:          "act-2",
						Description: "Negative action",
						Impact:      domain.ImpactNegative,
						Date:        time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
						Notes:       "",
					},
				},
			},
		},
	}

	// Render
	content := RenderMilestone(original)

	// Parse back
	parsed, err := ParseMilestone(content, "person-123")
	if err != nil {
		t.Fatalf("Round trip parse error: %v", err)
	}

	// Compare basic fields
	if parsed.ID != original.ID {
		t.Errorf("Round trip ID: got %v, want %v", parsed.ID, original.ID)
	}
	if parsed.Number != original.Number {
		t.Errorf("Round trip Number: got %v, want %v", parsed.Number, original.Number)
	}
	if parsed.Name != original.Name {
		t.Errorf("Round trip Name: got %v, want %v", parsed.Name, original.Name)
	}
	if parsed.Status != original.Status {
		t.Errorf("Round trip Status: got %v, want %v", parsed.Status, original.Status)
	}

	// Compare objectives
	if len(parsed.Objectives) != len(original.Objectives) {
		t.Fatalf("Round trip Objectives count: got %v, want %v",
			len(parsed.Objectives), len(original.Objectives))
	}

	parsedObj := parsed.Objectives[0]
	originalObj := original.Objectives[0]

	if parsedObj.ID != originalObj.ID {
		t.Errorf("Round trip Objective.ID: got %v, want %v", parsedObj.ID, originalObj.ID)
	}
	if parsedObj.Achieved != originalObj.Achieved {
		t.Errorf("Round trip Objective.Achieved: got %v, want %v",
			parsedObj.Achieved, originalObj.Achieved)
	}

	// Compare actions
	if len(parsedObj.Actions) != len(originalObj.Actions) {
		t.Fatalf("Round trip Actions count: got %v, want %v",
			len(parsedObj.Actions), len(originalObj.Actions))
	}

	for i := range originalObj.Actions {
		if parsedObj.Actions[i].Impact != originalObj.Actions[i].Impact {
			t.Errorf("Round trip Action[%d].Impact: got %v, want %v",
				i, parsedObj.Actions[i].Impact, originalObj.Actions[i].Impact)
		}
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantFm        string
		wantBody      string
		wantErr       bool
	}{
		{
			name: "valid frontmatter",
			content: `---
key: value
---

Body content`,
			wantFm:   "key: value",
			wantBody: "\nBody content",
			wantErr:  false,
		},
		{
			name:    "no frontmatter",
			content: "Just body content",
			wantErr: true,
		},
		{
			name: "unclosed frontmatter",
			content: `---
key: value
No closing marker`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := splitFrontmatter([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("splitFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if string(fm) != tt.wantFm {
					t.Errorf("splitFrontmatter() fm = %v, want %v", string(fm), tt.wantFm)
				}
				if string(body) != tt.wantBody {
					t.Errorf("splitFrontmatter() body = %v, want %v", string(body), tt.wantBody)
				}
			}
		})
	}
}
