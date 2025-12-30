package domain

import "errors"

// Validation errors.
var (
	ErrMissingID              = errors.New("id is required")
	ErrMissingName            = errors.New("name is required")
	ErrMissingPersonID        = errors.New("person_id is required")
	ErrMissingDescription     = errors.New("description is required")
	ErrInvalidMilestoneNumber = errors.New("milestone number must be >= 1")
	ErrInvalidImpact          = errors.New("impact must be 'positive' or 'negative'")
)

// Storage errors.
var (
	ErrPersonNotFound    = errors.New("person not found")
	ErrMilestoneNotFound = errors.New("milestone not found")
	ErrObjectiveNotFound = errors.New("objective not found")
	ErrActionNotFound    = errors.New("action not found")
	ErrInvalidMarkdown   = errors.New("invalid markdown format")
	ErrInvalidFrontmatter = errors.New("invalid YAML frontmatter")
)

// Operation errors.
var (
	ErrDuplicatePerson    = errors.New("person with this name already exists")
	ErrDuplicateMilestone = errors.New("milestone number already exists for this person")
)
