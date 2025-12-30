// Package service provides the business logic layer for the career tracker.
package service

import (
	"time"

	"teamsAndDragons/internal/domain"
)

// TrackerService defines the business operations for career tracking.
type TrackerService interface {
	// Person operations
	ListPeople() ([]domain.Person, error)
	GetPerson(id string) (*domain.Person, error)
	CreatePerson(name, notes string) (*domain.Person, error)
	UpdatePerson(id, name, notes string) (*domain.Person, error)
	DeletePerson(id string) error

	// Milestone operations
	ListMilestones(personID string) ([]domain.Milestone, error)
	GetMilestone(id string) (*domain.Milestone, error)
	CreateMilestone(personID, name, description string) (*domain.Milestone, error)
	UpdateMilestone(id string, name, description string, status domain.MilestoneStatus) (*domain.Milestone, error)
	DeleteMilestone(id string) error

	// Objective operations
	AddObjective(milestoneID, name, description string) (*domain.Milestone, error)
	UpdateObjective(milestoneID, objectiveID, name, description string) (*domain.Milestone, error)
	DeleteObjective(milestoneID, objectiveID string) (*domain.Milestone, error)

	// Action operations
	AddAction(milestoneID, objectiveID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error)
	UpdateAction(milestoneID, objectiveID, actionID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error)
	DeleteAction(milestoneID, objectiveID, actionID string) (*domain.Milestone, error)

	// Review operations
	SetObjectiveAchieved(milestoneID, objectiveID string, achieved bool, reviewNotes string) (*domain.Milestone, error)
	GetObjectiveReview(milestoneID, objectiveID string) (*ObjectiveReview, error)

	// Export
	ExportPerson(personID, outputPath string, includeStyles bool) error
	ExportAll(outputDir string, includeStyles bool) error
}

// ObjectiveReview provides a summary of an objective's status.
type ObjectiveReview struct {
	Objective       domain.Objective
	PositiveActions []domain.Action
	NegativeActions []domain.Action
	PositiveCount   int
	NegativeCount   int
	Net             int // PositiveCount - NegativeCount
}
