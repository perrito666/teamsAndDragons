// Package domain defines the core types for the career progression tracker.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// MilestoneStatus represents the status of a milestone.
type MilestoneStatus string

const (
	MilestoneInProgress MilestoneStatus = "in_progress"
	MilestoneCompleted  MilestoneStatus = "completed"
)

// Impact represents whether an action contributed positively or negatively.
type Impact string

const (
	ImpactPositive Impact = "positive"
	ImpactNegative Impact = "negative"
)

// Person represents a software engineer being tracked.
type Person struct {
	ID        string
	Name      string
	Notes     string // Free-text notes about the person
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the Person has required fields.
func (p *Person) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrMissingName
	}
	return nil
}

// Slug returns a URL-safe slug for the person's name.
func (p *Person) Slug() string {
	return slugify(p.Name)
}

// Milestone represents a career progression step (e.g., "Senior 1 → Senior 2").
type Milestone struct {
	ID       string
	PersonID string
	// Number is a sequential representation of the order of the milestone
	Number int
	// Name is how we call this milestone, typically the role or transition to be achieved "Senior 1 → Senior 2"
	Name string
	// Description contains a free text explanation of the milestione
	Description string
	// Status represents were this milestone is in the "workflow" (ie. In progress, complete)
	Status     MilestoneStatus
	Objectives []Objective
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks if the Milestone has required fields.
func (m *Milestone) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(m.PersonID) == "" {
		return ErrMissingPersonID
	}
	if strings.TrimSpace(m.Name) == "" {
		return ErrMissingName
	}
	if m.Number < 1 {
		return ErrInvalidMilestoneNumber
	}
	return nil
}

// Slug returns a URL-safe slug for the milestone name.
func (m *Milestone) Slug() string {
	return slugify(m.Name)
}

// Filename returns the expected filename for this milestone.
// Format: 01-milestone-{slug}.md
func (m *Milestone) Filename() string {
	return formatMilestoneFilename(m.Number, m.Slug())
}

// PositiveActions returns all actions with positive impact for this milestone.
func (m *Milestone) PositiveActions() []Action {
	var actions []Action
	for _, obj := range m.Objectives {
		for _, act := range obj.Actions {
			if act.Impact == ImpactPositive {
				actions = append(actions, act)
			}
		}
	}
	return actions
}

// NegativeActions returns all actions with negative impact for this milestone.
func (m *Milestone) NegativeActions() []Action {
	var actions []Action
	for _, obj := range m.Objectives {
		for _, act := range obj.Actions {
			if act.Impact == ImpactNegative {
				actions = append(actions, act)
			}
		}
	}
	return actions
}

// AchievedObjectives returns all objectives marked as achieved.
func (m *Milestone) AchievedObjectives() []Objective {
	var achieved []Objective
	for _, obj := range m.Objectives {
		if obj.Achieved {
			achieved = append(achieved, obj)
		}
	}
	return achieved
}

// PendingObjectives returns all objectives not yet achieved.
func (m *Milestone) PendingObjectives() []Objective {
	var pending []Objective
	for _, obj := range m.Objectives {
		if !obj.Achieved {
			pending = append(pending, obj)
		}
	}
	return pending
}

// Progress returns the percentage of achieved objectives (0-100).
func (m *Milestone) Progress() int {
	if len(m.Objectives) == 0 {
		return 0
	}
	achieved := len(m.AchievedObjectives())
	return (achieved * 100) / len(m.Objectives)
}

// Objective represents a goal within a milestone.
type Objective struct {
	ID          string
	Name        string // e.g., "Better communication of knowledge"
	Description string
	Achieved    bool
	ReviewNotes string // Optional notes when marking achieved/not achieved
	Actions     []Action
}

// Validate checks if the Objective has required fields.
func (o *Objective) Validate() error {
	if strings.TrimSpace(o.ID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(o.Name) == "" {
		return ErrMissingName
	}
	return nil
}

// PositiveActions returns all actions with positive impact.
func (o *Objective) PositiveActions() []Action {
	var actions []Action
	for _, act := range o.Actions {
		if act.Impact == ImpactPositive {
			actions = append(actions, act)
		}
	}
	return actions
}

// NegativeActions returns all actions with negative impact.
func (o *Objective) NegativeActions() []Action {
	var actions []Action
	for _, act := range o.Actions {
		if act.Impact == ImpactNegative {
			actions = append(actions, act)
		}
	}
	return actions
}

// Tally returns the count of positive and negative actions.
func (o *Objective) Tally() (positive, negative int) {
	for _, act := range o.Actions {
		if act.Impact == ImpactPositive {
			positive++
		} else {
			negative++
		}
	}
	return
}

// Action represents something that contributes to an objective.
type Action struct {
	ID          string
	Description string
	Impact      Impact // positive or negative
	Date        time.Time
	Notes       string // Additional context
}

// Validate checks if the Action has required fields.
func (a *Action) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(a.Description) == "" {
		return ErrMissingDescription
	}
	if a.Impact != ImpactPositive && a.Impact != ImpactNegative {
		return ErrInvalidImpact
	}
	return nil
}

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " → ", "-to-")
	s = strings.ReplaceAll(s, "→", "-to-")

	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteRune('-')
		}
	}

	// Remove consecutive dashes
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// remove trailing dashes
	slug = strings.Trim(slug, "-")

	return slug
}

// formatMilestoneFilename creates a milestone filename.
// Format: 01-milestone-{slug}.md
func formatMilestoneFilename(number int, slug string) string {
	return fmt.Sprintf("%02d-milestone-%s.md", number, slug)
}
