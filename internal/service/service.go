// Package service provides the business logic layer for the career tracker.
package service

import (
	"time"

	"teamsAndDragons/internal/domain"
)

// TrackerService defines the business operations for career tracking.
type TrackerService interface {
	// Person operations

	// ListPeople returns a list of the people in this tracked team (in this folder)
	ListPeople() ([]domain.Person, error)
	// GetPerson returns a single person tracked in this folder.
	GetPerson(id string) (*domain.Person, error)
	// CreatePerson adds a tracked person to this folder
	CreatePerson(name, notes string) (*domain.Person, error)
	// UpdatePerson replaces existing name and notes with the passed in an existing person (or fails)
	UpdatePerson(id, name, notes string) (*domain.Person, error)
	// DeletePerson deletes the person with the given id if it exists or fails
	DeletePerson(id string) error

	// Milestone operations

	// ListMilestones returns a list of the available Milestones for a given person
	ListMilestones(personID string) ([]domain.Milestone, error)
	// GetMilestone returns a single milestone by ID.
	GetMilestone(personID, id string) (*domain.Milestone, error)
	// CreateMilestone Adds a new Milestone to the given person.
	CreateMilestone(personID, name, description string) (*domain.Milestone, error)
	// UpdateMilestone updates the name, description and status of a given milestone.
	UpdateMilestone(personID, id string, name, description string, status domain.MilestoneStatus) (*domain.Milestone, error)
	// DeleteMilestone deletes a given milestone by ID.
	DeleteMilestone(personID, id string) error

	// Objective operations

	// AddObjective creates a new Objective within the given milestone
	AddObjective(personID, milestoneID, name, description string) (*domain.Milestone, error)
	// UpdateObjective changes the name and description of the passed objective within the passed milestones
	UpdateObjective(personID, milestoneID, objectiveID, name, description string) (*domain.Milestone, error)
	// DeleteObjective deletes the given objective within the given miletone.
	DeleteObjective(personID, milestoneID, objectiveID string) (*domain.Milestone, error)

	// Action operations

	// AddAction Adds an action towards/against an objective within the given objective.
	AddAction(personID, milestoneID, objectiveID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error)
	// UpdateAction updates all the fields of an Action (it is basically Add all over again)
	UpdateAction(personID, milestoneID, objectiveID, actionID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error)
	// DeleteAction deletes the given action
	DeleteAction(personID, milestoneID, objectiveID, actionID string) (*domain.Milestone, error)

	// Review operations

	// SetObjectiveAchieved sets the achievement status of the given objective.
	SetObjectiveAchieved(personID, milestoneID, objectiveID string, achieved bool, reviewNotes string) (*domain.Milestone, error)
	// GetObjectiveReview returns an object containing an overview of the given objective status.
	GetObjectiveReview(personID, milestoneID, objectiveID string) (*ObjectiveReview, error)

	// Export

	// ExportPerson generates an html one pager of the person's progress
	ExportPerson(personID, outputPath string, includeStyles bool) error
	// ExportAll generates htmls for all the people tracked
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
