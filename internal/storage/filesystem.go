package storage

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"teamsAndDragons/internal/domain"

	"github.com/google/uuid"
)

const (
	profileFilename = "00-profile.md"
)

// FileStorage implements Storage using the filesystem with markdown files.
type FileStorage struct {
	dataDir string
	logger  *slog.Logger
}

// NewFileStorage creates a new FileStorage instance.
func NewFileStorage(dataDir string, logger *slog.Logger) (*FileStorage, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
		logger:  logger,
	}, nil
}

// ListPeople returns all tracked people.
func (fs *FileStorage) ListPeople() ([]domain.Person, error) {
	entries, err := os.ReadDir(fs.dataDir)
	if err != nil {
		return nil, fmt.Errorf("reading data directory: %w", err)
	}

	var people []domain.Person
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		profilePath := filepath.Join(fs.dataDir, entry.Name(), profileFilename)
		content, err := os.ReadFile(profilePath)
		if err != nil {
			fs.logger.Warn("skipping directory without profile", "dir", entry.Name(), "error", err)
			continue
		}

		person, err := ParseProfile(content)
		if err != nil {
			fs.logger.Warn("skipping invalid profile", "dir", entry.Name(), "error", err)
			continue
		}

		people = append(people, *person)
	}

	// Sort by name
	sort.Slice(people, func(i, j int) bool {
		return people[i].Name < people[j].Name
	})

	return people, nil
}

// GetPerson retrieves a person by ID.
func (fs *FileStorage) GetPerson(id string) (*domain.Person, error) {
	people, err := fs.ListPeople()
	if err != nil {
		return nil, err
	}

	for _, p := range people {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, domain.ErrPersonNotFound
}

// GetPersonBySlug retrieves a person by their folder name.
func (fs *FileStorage) GetPersonBySlug(slug string) (*domain.Person, error) {
	profilePath := filepath.Join(fs.dataDir, slug, profileFilename)
	content, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrPersonNotFound
		}
		return nil, fmt.Errorf("reading profile: %w", err)
	}

	return ParseProfile(content)
}

// SavePerson creates or updates a person.
func (fs *FileStorage) SavePerson(person *domain.Person) error {
	// Generate ID if new
	if person.ID == "" {
		person.ID = uuid.New().String()
		person.CreatedAt = time.Now()
	}
	person.UpdatedAt = time.Now()

	if err := person.Validate(); err != nil {
		return err
	}

	// Check for duplicates (different ID, same slug)
	slug := person.Slug()
	existing, err := fs.GetPersonBySlug(slug)
	if err == nil && existing.ID != person.ID {
		return domain.ErrDuplicatePerson
	}

	// Create person directory
	personDir := filepath.Join(fs.dataDir, slug)
	if err := os.MkdirAll(personDir, 0755); err != nil {
		return fmt.Errorf("creating person directory: %w", err)
	}

	// Write profile
	content := RenderProfile(person)
	profilePath := filepath.Join(personDir, profileFilename)
	if err := os.WriteFile(profilePath, content, 0644); err != nil {
		return fmt.Errorf("writing profile: %w", err)
	}

	fs.logger.Info("saved person", "id", person.ID, "name", person.Name)
	return nil
}

// DeletePerson removes a person and all their data.
func (fs *FileStorage) DeletePerson(id string) error {
	person, err := fs.GetPerson(id)
	if err != nil {
		return err
	}

	personDir := filepath.Join(fs.dataDir, person.Slug())
	if err := os.RemoveAll(personDir); err != nil {
		return fmt.Errorf("removing person directory: %w", err)
	}

	fs.logger.Info("deleted person", "id", id, "name", person.Name)
	return nil
}

// ListMilestones returns all milestones for a person.
func (fs *FileStorage) ListMilestones(personID string) ([]domain.Milestone, error) {
	person, err := fs.GetPerson(personID)
	if err != nil {
		return nil, err
	}

	personDir := filepath.Join(fs.dataDir, person.Slug())
	entries, err := os.ReadDir(personDir)
	if err != nil {
		return nil, fmt.Errorf("reading person directory: %w", err)
	}

	milestonePattern := regexp.MustCompile(`^(\d+)-milestone-.*\.md$`)
	var milestones []domain.Milestone

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !milestonePattern.MatchString(entry.Name()) {
			continue
		}

		milestonePath := filepath.Join(personDir, entry.Name())
		content, err := os.ReadFile(milestonePath)
		if err != nil {
			fs.logger.Warn("skipping unreadable milestone", "file", entry.Name(), "error", err)
			continue
		}

		milestone, err := ParseMilestone(content, personID)
		if err != nil {
			fs.logger.Warn("skipping invalid milestone", "file", entry.Name(), "error", err)
			continue
		}

		milestones = append(milestones, *milestone)
	}

	// Sort by number
	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].Number < milestones[j].Number
	})

	return milestones, nil
}

// GetMilestone retrieves a milestone by ID.
func (fs *FileStorage) GetMilestone(id string) (*domain.Milestone, error) {
	people, err := fs.ListPeople()
	if err != nil {
		return nil, err
	}

	for _, person := range people {
		milestones, err := fs.ListMilestones(person.ID)
		if err != nil {
			continue
		}

		for _, m := range milestones {
			if m.ID == id {
				return &m, nil
			}
		}
	}

	return nil, domain.ErrMilestoneNotFound
}

// SaveMilestone creates or updates a milestone.
func (fs *FileStorage) SaveMilestone(milestone *domain.Milestone) error {
	// Generate ID if new
	if milestone.ID == "" {
		milestone.ID = uuid.New().String()
		milestone.CreatedAt = time.Now()
	}
	milestone.UpdatedAt = time.Now()

	// Default status
	if milestone.Status == "" {
		milestone.Status = domain.MilestoneInProgress
	}

	if err := milestone.Validate(); err != nil {
		return err
	}

	// Get person for directory
	person, err := fs.GetPerson(milestone.PersonID)
	if err != nil {
		return err
	}

	// Generate IDs for objectives and actions if needed
	for i := range milestone.Objectives {
		if milestone.Objectives[i].ID == "" {
			milestone.Objectives[i].ID = fmt.Sprintf("obj-%s", uuid.New().String()[:8])
		}
		for j := range milestone.Objectives[i].Actions {
			if milestone.Objectives[i].Actions[j].ID == "" {
				milestone.Objectives[i].Actions[j].ID = fmt.Sprintf("act-%s", uuid.New().String()[:8])
			}
		}
	}

	// Find and remove old file if milestone number changed
	personDir := filepath.Join(fs.dataDir, person.Slug())
	existingMilestones, _ := fs.ListMilestones(person.ID)
	for _, existing := range existingMilestones {
		if existing.ID == milestone.ID && existing.Number != milestone.Number {
			oldPath := filepath.Join(personDir, existing.Filename())
			_ = os.Remove(oldPath)
			break
		}
	}

	// Write milestone file
	content := RenderMilestone(milestone)
	milestonePath := filepath.Join(personDir, milestone.Filename())
	if err := os.WriteFile(milestonePath, content, 0644); err != nil {
		return fmt.Errorf("writing milestone: %w", err)
	}

	fs.logger.Info("saved milestone", "id", milestone.ID, "name", milestone.Name, "person", person.Name)
	return nil
}

// DeleteMilestone removes a milestone.
func (fs *FileStorage) DeleteMilestone(id string) error {
	milestone, err := fs.GetMilestone(id)
	if err != nil {
		return err
	}

	person, err := fs.GetPerson(milestone.PersonID)
	if err != nil {
		return err
	}

	milestonePath := filepath.Join(fs.dataDir, person.Slug(), milestone.Filename())
	if err := os.Remove(milestonePath); err != nil {
		return fmt.Errorf("removing milestone file: %w", err)
	}

	fs.logger.Info("deleted milestone", "id", id, "name", milestone.Name)
	return nil
}

// GetNextMilestoneNumber returns the next available milestone number for a person.
func (fs *FileStorage) GetNextMilestoneNumber(personID string) (int, error) {
	milestones, err := fs.ListMilestones(personID)
	if err != nil {
		return 1, nil
	}

	maxNumber := 0
	for _, m := range milestones {
		if m.Number > maxNumber {
			maxNumber = m.Number
		}
	}

	return maxNumber + 1, nil
}

// personDirPath returns the directory path for a person.
func (fs *FileStorage) personDirPath(slug string) string {
	return filepath.Join(fs.dataDir, slug)
}

// RenamePersonDir renames a person's directory when their name changes.
func (fs *FileStorage) RenamePersonDir(oldSlug, newSlug string) error {
	if oldSlug == newSlug {
		return nil
	}

	oldPath := filepath.Join(fs.dataDir, oldSlug)
	newPath := filepath.Join(fs.dataDir, newSlug)

	// Check if old directory exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil // Nothing to rename
	}

	// Check if new directory already exists
	if _, err := os.Stat(newPath); err == nil {
		return domain.ErrDuplicatePerson
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("renaming directory: %w", err)
	}

	return nil
}

// Ensure FileStorage implements Storage interface.
var _ Storage = (*FileStorage)(nil)
