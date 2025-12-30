// Package storage provides markdown-based persistence for the career tracker.
package storage

import (
	"teamsAndDragons/internal/domain"
)

// Storage defines the interface for persisting tracker data.
type Storage interface {
	PersonStorage
	MilestoneStorage
}

// PersonStorage defines operations for Person entities.
type PersonStorage interface {
	// ListPeople returns all tracked people.
	ListPeople() ([]domain.Person, error)

	// GetPerson retrieves a person by ID.
	GetPerson(id string) (*domain.Person, error)

	// GetPersonBySlug retrieves a person by their name slug (folder name).
	GetPersonBySlug(slug string) (*domain.Person, error)

	// SavePerson creates or updates a person.
	SavePerson(person *domain.Person) error

	// DeletePerson removes a person and all their data.
	DeletePerson(id string) error
}

// MilestoneStorage defines operations for Milestone entities.
type MilestoneStorage interface {
	// ListMilestones returns all milestones for a person.
	ListMilestones(personID string) ([]domain.Milestone, error)

	// GetMilestone retrieves a milestone by ID.
	GetMilestone(id string) (*domain.Milestone, error)

	// SaveMilestone creates or updates a milestone.
	SaveMilestone(milestone *domain.Milestone) error

	// DeleteMilestone removes a milestone.
	DeleteMilestone(id string) error
}

// ExportOptions configures HTML export behavior.
type ExportOptions struct {
	IncludeStyles bool   // Embed CSS styles
	OutputPath    string // Destination file path
}

// Exporter defines HTML export functionality.
type Exporter interface {
	// ExportPerson exports a single person's data to HTML.
	ExportPerson(personID string, opts ExportOptions) error

	// ExportAll exports all people to HTML.
	ExportAll(opts ExportOptions) error
}
