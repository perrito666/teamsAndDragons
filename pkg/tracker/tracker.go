package tracker

import (
	"time"

	"teamsAndDragons/internal/domain"
	"teamsAndDragons/internal/service"
	"teamsAndDragons/internal/storage"
)

// Re-export domain types for public use.
type (
	Person          = domain.Person
	Milestone       = domain.Milestone
	MilestoneStatus = domain.MilestoneStatus
	Objective       = domain.Objective
	Action          = domain.Action
	Impact          = domain.Impact
	ObjectiveReview = service.ObjectiveReview
)

// Re-export constants.
const (
	MilestoneInProgress = domain.MilestoneInProgress
	MilestoneCompleted  = domain.MilestoneCompleted
	ImpactPositive      = domain.ImpactPositive
	ImpactNegative      = domain.ImpactNegative
)

// Re-export errors.
var (
	ErrPersonNotFound    = domain.ErrPersonNotFound
	ErrMilestoneNotFound = domain.ErrMilestoneNotFound
	ErrObjectiveNotFound = domain.ErrObjectiveNotFound
	ErrActionNotFound    = domain.ErrActionNotFound
	ErrDuplicatePerson   = domain.ErrDuplicatePerson
)

// Tracker provides the public API for career progression tracking.
type Tracker struct {
	svc service.TrackerService
}

// New creates a new Tracker with the given options.
func New(opts ...Option) (*Tracker, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	store, err := storage.NewFileStorage(options.DataDir, options.Logger)
	if err != nil {
		return nil, err
	}

	exporter := storage.NewHTMLExporter(store)
	svc := service.NewTracker(store, exporter, options.Logger)

	return &Tracker{svc: svc}, nil
}

// --- Person Operations ---

// ListPeople returns all tracked people.
func (t *Tracker) ListPeople() ([]Person, error) {
	return t.svc.ListPeople()
}

// GetPerson retrieves a person by ID.
func (t *Tracker) GetPerson(id string) (*Person, error) {
	return t.svc.GetPerson(id)
}

// CreatePerson creates a new person to track.
func (t *Tracker) CreatePerson(name, notes string) (*Person, error) {
	return t.svc.CreatePerson(name, notes)
}

// UpdatePerson updates an existing person.
func (t *Tracker) UpdatePerson(id, name, notes string) (*Person, error) {
	return t.svc.UpdatePerson(id, name, notes)
}

// DeletePerson removes a person and all their data.
func (t *Tracker) DeletePerson(id string) error {
	return t.svc.DeletePerson(id)
}

// --- Milestone Operations ---

// ListMilestones returns all milestones for a person.
func (t *Tracker) ListMilestones(personID string) ([]Milestone, error) {
	return t.svc.ListMilestones(personID)
}

// GetMilestone retrieves a milestone by ID.
func (t *Tracker) GetMilestone(personID, id string) (*Milestone, error) {
	return t.svc.GetMilestone(personID, id)
}

// CreateMilestone creates a new milestone for a person.
func (t *Tracker) CreateMilestone(personID, name, description string) (*Milestone, error) {
	return t.svc.CreateMilestone(personID, name, description)
}

// UpdateMilestone updates an existing milestone.
func (t *Tracker) UpdateMilestone(personID, id, name, description string, status MilestoneStatus) (*Milestone, error) {
	return t.svc.UpdateMilestone(personID, id, name, description, status)
}

// DeleteMilestone removes a milestone.
func (t *Tracker) DeleteMilestone(personID, id string) error {
	return t.svc.DeleteMilestone(personID, id)
}

// --- Objective Operations ---

// AddObjective adds an objective to a milestone.
func (t *Tracker) AddObjective(personID, milestoneID, name, description string) (*Milestone, error) {
	return t.svc.AddObjective(personID, milestoneID, name, description)
}

// UpdateObjective updates an objective.
func (t *Tracker) UpdateObjective(personID, milestoneID, objectiveID, name, description string) (*Milestone, error) {
	return t.svc.UpdateObjective(personID, milestoneID, objectiveID, name, description)
}

// DeleteObjective removes an objective.
func (t *Tracker) DeleteObjective(personID, milestoneID, objectiveID string) (*Milestone, error) {
	return t.svc.DeleteObjective(personID, milestoneID, objectiveID)
}

// --- Action Operations ---

// AddAction adds an action to an objective.
func (t *Tracker) AddAction(personID, milestoneID, objectiveID, description string, impact Impact, date time.Time, notes string) (*Milestone, error) {
	return t.svc.AddAction(personID, milestoneID, objectiveID, description, impact, date, notes)
}

// UpdateAction updates an action.
func (t *Tracker) UpdateAction(personID, milestoneID, objectiveID, actionID, description string, impact Impact, date time.Time, notes string) (*Milestone, error) {
	return t.svc.UpdateAction(personID, milestoneID, objectiveID, actionID, description, impact, date, notes)
}

// DeleteAction removes an action.
func (t *Tracker) DeleteAction(personID, milestoneID, objectiveID, actionID string) (*Milestone, error) {
	return t.svc.DeleteAction(personID, milestoneID, objectiveID, actionID)
}

// --- Review Operations ---

// SetObjectiveAchieved marks an objective as achieved or not achieved.
func (t *Tracker) SetObjectiveAchieved(personID, milestoneID, objectiveID string, achieved bool, reviewNotes string) (*Milestone, error) {
	return t.svc.SetObjectiveAchieved(personID, milestoneID, objectiveID, achieved, reviewNotes)
}

// GetObjectiveReview returns a review summary for an objective.
func (t *Tracker) GetObjectiveReview(personID, milestoneID, objectiveID string) (*ObjectiveReview, error) {
	return t.svc.GetObjectiveReview(personID, milestoneID, objectiveID)
}

// --- Export Operations ---

// ExportPerson exports a person's data to HTML.
func (t *Tracker) ExportPerson(personID, outputPath string) error {
	return t.svc.ExportPerson(personID, outputPath, true)
}

// ExportAll exports all people to HTML files in a directory.
func (t *Tracker) ExportAll(outputDir string) error {
	return t.svc.ExportAll(outputDir, true)
}
