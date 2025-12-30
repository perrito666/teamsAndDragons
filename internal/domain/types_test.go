package domain

import (
	"testing"
	"time"
)

func TestPerson_Validate(t *testing.T) {
	tests := []struct {
		name    string
		person  Person
		wantErr error
	}{
		{
			name: "valid person",
			person: Person{
				ID:   "123",
				Name: "John Doe",
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			person: Person{
				Name: "John Doe",
			},
			wantErr: ErrMissingID,
		},
		{
			name: "missing name",
			person: Person{
				ID: "123",
			},
			wantErr: ErrMissingName,
		},
		{
			name: "whitespace only ID",
			person: Person{
				ID:   "   ",
				Name: "John Doe",
			},
			wantErr: ErrMissingID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.person.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPerson_Slug(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		expected string
	}{
		{
			name:     "simple name",
			person:   Person{Name: "John Doe"},
			expected: "john-doe",
		},
		{
			name:     "name with special chars",
			person:   Person{Name: "María García"},
			expected: "mara-garca",
		},
		{
			name:     "uppercase",
			person:   Person{Name: "ALICE SMITH"},
			expected: "alice-smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.person.Slug()
			if got != tt.expected {
				t.Errorf("Slug() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMilestone_Validate(t *testing.T) {
	tests := []struct {
		name      string
		milestone Milestone
		wantErr   error
	}{
		{
			name: "valid milestone",
			milestone: Milestone{
				ID:       "123",
				PersonID: "456",
				Name:     "Senior 1 → Senior 2",
				Number:   1,
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			milestone: Milestone{
				PersonID: "456",
				Name:     "Test",
				Number:   1,
			},
			wantErr: ErrMissingID,
		},
		{
			name: "missing person ID",
			milestone: Milestone{
				ID:     "123",
				Name:   "Test",
				Number: 1,
			},
			wantErr: ErrMissingPersonID,
		},
		{
			name: "invalid number",
			milestone: Milestone{
				ID:       "123",
				PersonID: "456",
				Name:     "Test",
				Number:   0,
			},
			wantErr: ErrInvalidMilestoneNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.milestone.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMilestone_Slug(t *testing.T) {
	tests := []struct {
		name      string
		milestone Milestone
		expected  string
	}{
		{
			name:      "arrow notation",
			milestone: Milestone{Name: "Senior 1 → Senior 2"},
			expected:  "senior-1-to-senior-2",
		},
		{
			name:      "simple name",
			milestone: Milestone{Name: "Staff Engineer"},
			expected:  "staff-engineer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.milestone.Slug()
			if got != tt.expected {
				t.Errorf("Slug() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMilestone_Filename(t *testing.T) {
	ms := Milestone{
		Number: 1,
		Name:   "Senior 1 → Senior 2",
	}

	expected := "01-milestone-senior-1-to-senior-2.md"
	got := ms.Filename()

	if got != expected {
		t.Errorf("Filename() = %v, want %v", got, expected)
	}
}

func TestMilestone_Progress(t *testing.T) {
	tests := []struct {
		name       string
		objectives []Objective
		expected   int
	}{
		{
			name:       "no objectives",
			objectives: nil,
			expected:   0,
		},
		{
			name: "none achieved",
			objectives: []Objective{
				{Achieved: false},
				{Achieved: false},
			},
			expected: 0,
		},
		{
			name: "half achieved",
			objectives: []Objective{
				{Achieved: true},
				{Achieved: false},
			},
			expected: 50,
		},
		{
			name: "all achieved",
			objectives: []Objective{
				{Achieved: true},
				{Achieved: true},
			},
			expected: 100,
		},
		{
			name: "one of three",
			objectives: []Objective{
				{Achieved: true},
				{Achieved: false},
				{Achieved: false},
			},
			expected: 33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := Milestone{Objectives: tt.objectives}
			got := ms.Progress()
			if got != tt.expected {
				t.Errorf("Progress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestObjective_Tally(t *testing.T) {
	obj := Objective{
		Actions: []Action{
			{Impact: ImpactPositive},
			{Impact: ImpactPositive},
			{Impact: ImpactNegative},
			{Impact: ImpactPositive},
		},
	}

	pos, neg := obj.Tally()

	if pos != 3 {
		t.Errorf("Tally() positive = %v, want 3", pos)
	}
	if neg != 1 {
		t.Errorf("Tally() negative = %v, want 1", neg)
	}
}

func TestAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		action  Action
		wantErr error
	}{
		{
			name: "valid action",
			action: Action{
				ID:          "123",
				Description: "Led team workshop",
				Impact:      ImpactPositive,
				Date:        time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			action: Action{
				Description: "Test",
				Impact:      ImpactPositive,
			},
			wantErr: ErrMissingID,
		},
		{
			name: "missing description",
			action: Action{
				ID:     "123",
				Impact: ImpactPositive,
			},
			wantErr: ErrMissingDescription,
		},
		{
			name: "invalid impact",
			action: Action{
				ID:          "123",
				Description: "Test",
				Impact:      "invalid",
			},
			wantErr: ErrInvalidImpact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Senior 1 → Senior 2", "senior-1-to-senior-2"},
		{"Test  Multiple   Spaces", "test-multiple-spaces"},
		{"Special!@#$%Chars", "specialchars"},
		{"Already-Slugified", "already-slugified"},
		{"   Trim Spaces   ", "trim-spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
