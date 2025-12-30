package service

import (
	"fmt"
	"log/slog"
	"time"

	"teamsAndDragons/internal/domain"
	"teamsAndDragons/internal/storage"
)

// Tracker implements TrackerService.
type Tracker struct {
	storage  storage.Storage
	exporter storage.Exporter
	logger   *slog.Logger
}

// NewTracker creates a new Tracker service.
func NewTracker(store storage.Storage, exporter storage.Exporter, logger *slog.Logger) *Tracker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Tracker{
		storage:  store,
		exporter: exporter,
		logger:   logger,
	}
}

// ListPeople returns all tracked people.
func (t *Tracker) ListPeople() ([]domain.Person, error) {
	return t.storage.ListPeople()
}

// GetPerson retrieves a person by ID.
func (t *Tracker) GetPerson(id string) (*domain.Person, error) {
	return t.storage.GetPerson(id)
}

// CreatePerson creates a new person.
func (t *Tracker) CreatePerson(name, notes string) (*domain.Person, error) {
	person := &domain.Person{
		Name:  name,
		Notes: notes,
	}

	if err := t.storage.SavePerson(person); err != nil {
		return nil, err
	}

	t.logger.Debug("created person", "id", person.ID, "name", name)
	return person, nil
}

// UpdatePerson updates an existing person.
func (t *Tracker) UpdatePerson(id, name, notes string) (*domain.Person, error) {
	person, err := t.storage.GetPerson(id)
	if err != nil {
		return nil, err
	}

	// Track if we need to rename directory
	oldSlug := person.Slug()

	person.Name = name
	person.Notes = notes

	newSlug := person.Slug()

	// Handle directory rename if name changed
	if oldSlug != newSlug {
		if fs, ok := t.storage.(*storage.FileStorage); ok {
			if err := fs.RenamePersonDir(oldSlug, newSlug); err != nil {
				return nil, fmt.Errorf("renaming person directory: %w", err)
			}
		}
	}

	if err := t.storage.SavePerson(person); err != nil {
		return nil, err
	}

	t.logger.Debug("updated person", "id", id, "name", name)
	return person, nil
}

// DeletePerson removes a person.
func (t *Tracker) DeletePerson(id string) error {
	return t.storage.DeletePerson(id)
}

// ListMilestones returns all milestones for a person.
func (t *Tracker) ListMilestones(personID string) ([]domain.Milestone, error) {
	return t.storage.ListMilestones(personID)
}

// GetMilestone retrieves a milestone by ID.
func (t *Tracker) GetMilestone(personID, id string) (*domain.Milestone, error) {
	return t.storage.GetMilestone(personID, id)
}

// CreateMilestone creates a new milestone.
func (t *Tracker) CreateMilestone(personID, name, description string) (*domain.Milestone, error) {
	// Verify person exists
	if _, err := t.storage.GetPerson(personID); err != nil {
		return nil, err
	}

	// Get next milestone number
	fs, ok := t.storage.(*storage.FileStorage)
	if !ok {
		return nil, fmt.Errorf("storage type does not support milestone numbering")
	}

	number, err := fs.GetNextMilestoneNumber(personID)
	if err != nil {
		return nil, err
	}

	milestone := &domain.Milestone{
		PersonID:    personID,
		Number:      number,
		Name:        name,
		Description: description,
		Status:      domain.MilestoneInProgress,
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("created milestone", "id", milestone.ID, "name", name, "number", number)
	return milestone, nil
}

// UpdateMilestone updates an existing milestone.
func (t *Tracker) UpdateMilestone(personID, id string, name, description string, status domain.MilestoneStatus) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, id)
	if err != nil {
		return nil, err
	}

	milestone.Name = name
	milestone.Description = description
	milestone.Status = status

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("updated milestone", "id", id, "name", name)
	return milestone, nil
}

// DeleteMilestone removes a milestone.
func (t *Tracker) DeleteMilestone(personID, id string) error {
	return t.storage.DeleteMilestone(personID, id)
}

// AddObjective adds an objective to a milestone.
func (t *Tracker) AddObjective(personID, milestoneID, name, description string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	objective := domain.Objective{
		Name:        name,
		Description: description,
		Achieved:    false,
	}

	milestone.Objectives = append(milestone.Objectives, objective)

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("added objective", "milestone_id", milestoneID, "name", name)
	return milestone, nil
}

// UpdateObjective updates an objective.
func (t *Tracker) UpdateObjective(personID, milestoneID, objectiveID, name, description string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID == objectiveID {
			milestone.Objectives[i].Name = name
			milestone.Objectives[i].Description = description
			found = true
			break
		}
	}

	if !found {
		return nil, domain.ErrObjectiveNotFound
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("updated objective", "milestone_id", milestoneID, "objective_id", objectiveID)
	return milestone, nil
}

// DeleteObjective removes an objective.
func (t *Tracker) DeleteObjective(personID, milestoneID, objectiveID string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	found := false
	objectives := make([]domain.Objective, 0, len(milestone.Objectives)-1)
	for _, obj := range milestone.Objectives {
		if obj.ID == objectiveID {
			found = true
			continue
		}
		objectives = append(objectives, obj)
	}

	if !found {
		return nil, domain.ErrObjectiveNotFound
	}

	milestone.Objectives = objectives

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("deleted objective", "milestone_id", milestoneID, "objective_id", objectiveID)
	return milestone, nil
}

// AddAction adds an action to an objective.
func (t *Tracker) AddAction(personID, milestoneID, objectiveID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	action := domain.Action{
		Description: description,
		Impact:      impact,
		Date:        date,
		Notes:       notes,
	}

	found := false
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID == objectiveID {
			milestone.Objectives[i].Actions = append(milestone.Objectives[i].Actions, action)
			found = true
			break
		}
	}

	if !found {
		return nil, domain.ErrObjectiveNotFound
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("added action", "milestone_id", milestoneID, "objective_id", objectiveID, "impact", impact)
	return milestone, nil
}

// UpdateAction updates an action.
func (t *Tracker) UpdateAction(personID, milestoneID, objectiveID, actionID, description string, impact domain.Impact, date time.Time, notes string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID != objectiveID {
			continue
		}
		for j := range milestone.Objectives[i].Actions {
			if milestone.Objectives[i].Actions[j].ID == actionID {
				milestone.Objectives[i].Actions[j].Description = description
				milestone.Objectives[i].Actions[j].Impact = impact
				milestone.Objectives[i].Actions[j].Date = date
				milestone.Objectives[i].Actions[j].Notes = notes
				found = true
				break
			}
		}
	}

	if !found {
		return nil, domain.ErrActionNotFound
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("updated action", "milestone_id", milestoneID, "action_id", actionID)
	return milestone, nil
}

// DeleteAction removes an action.
func (t *Tracker) DeleteAction(personID, milestoneID, objectiveID, actionID string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID != objectiveID {
			continue
		}
		actions := make([]domain.Action, 0, len(milestone.Objectives[i].Actions)-1)
		for _, act := range milestone.Objectives[i].Actions {
			if act.ID == actionID {
				found = true
				continue
			}
			actions = append(actions, act)
		}
		milestone.Objectives[i].Actions = actions
	}

	if !found {
		return nil, domain.ErrActionNotFound
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("deleted action", "milestone_id", milestoneID, "action_id", actionID)
	return milestone, nil
}

// SetObjectiveAchieved marks an objective as achieved or not achieved.
func (t *Tracker) SetObjectiveAchieved(personID, milestoneID, objectiveID string, achieved bool, reviewNotes string) (*domain.Milestone, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID == objectiveID {
			milestone.Objectives[i].Achieved = achieved
			if reviewNotes != "" {
				milestone.Objectives[i].ReviewNotes = reviewNotes
			}
			found = true
			break
		}
	}

	if !found {
		return nil, domain.ErrObjectiveNotFound
	}

	if err := t.storage.SaveMilestone(milestone); err != nil {
		return nil, err
	}

	t.logger.Debug("set objective achieved", "milestone_id", milestoneID, "objective_id", objectiveID, "achieved", achieved)
	return milestone, nil
}

// GetObjectiveReview returns a review summary for an objective.
func (t *Tracker) GetObjectiveReview(personID, milestoneID, objectiveID string) (*ObjectiveReview, error) {
	milestone, err := t.storage.GetMilestone(personID, milestoneID)
	if err != nil {
		return nil, err
	}

	for _, obj := range milestone.Objectives {
		if obj.ID == objectiveID {
			positiveActions := obj.PositiveActions()
			negativeActions := obj.NegativeActions()
			return &ObjectiveReview{
				Objective:       obj,
				PositiveActions: positiveActions,
				NegativeActions: negativeActions,
				PositiveCount:   len(positiveActions),
				NegativeCount:   len(negativeActions),
				Net:             len(positiveActions) - len(negativeActions),
			}, nil
		}
	}

	return nil, domain.ErrObjectiveNotFound
}

// ExportPerson exports a person to HTML.
func (t *Tracker) ExportPerson(personID, outputPath string, includeStyles bool) error {
	return t.exporter.ExportPerson(personID, storage.ExportOptions{
		OutputPath:    outputPath,
		IncludeStyles: includeStyles,
	})
}

// ExportAll exports all people to HTML.
func (t *Tracker) ExportAll(outputDir string, includeStyles bool) error {
	return t.exporter.ExportAll(storage.ExportOptions{
		OutputPath:    outputDir,
		IncludeStyles: includeStyles,
	})
}

// Ensure Tracker implements TrackerService.
var _ TrackerService = (*Tracker)(nil)
