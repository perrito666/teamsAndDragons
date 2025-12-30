package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"teamsAndDragons/internal/domain"

	"gopkg.in/yaml.v3"
)

// ProfileFrontmatter represents the YAML frontmatter in profile files.
type ProfileFrontmatter struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// MilestoneFrontmatter represents the YAML frontmatter in milestone files.
type MilestoneFrontmatter struct {
	ID        string                 `yaml:"id"`
	Number    int                    `yaml:"number"`
	Name      string                 `yaml:"name"`
	Status    domain.MilestoneStatus `yaml:"status"`
	CreatedAt time.Time              `yaml:"created_at"`
	UpdatedAt time.Time              `yaml:"updated_at"`
}

// ParseProfile parses a profile markdown file.
func ParseProfile(content []byte) (*domain.Person, error) {
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	var fm ProfileFrontmatter
	if err := yaml.Unmarshal(frontmatter, &fm); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidFrontmatter, err)
	}

	// Body is the free-text notes (skip the first heading line)
	notes := extractNotesFromBody(body)

	return &domain.Person{
		ID:        fm.ID,
		Name:      fm.Name,
		Notes:     notes,
		CreatedAt: fm.CreatedAt,
		UpdatedAt: fm.UpdatedAt,
	}, nil
}

// RenderProfile renders a person to markdown format.
func RenderProfile(person *domain.Person) []byte {
	var buf bytes.Buffer

	// Frontmatter
	fm := ProfileFrontmatter{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(fm)
	buf.WriteString("---\n\n")

	// Title
	buf.WriteString(fmt.Sprintf("# %s\n\n", person.Name))

	// Notes
	if person.Notes != "" {
		buf.WriteString(person.Notes)
		if !strings.HasSuffix(person.Notes, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes()
}

// ParseMilestone parses a milestone markdown file.
func ParseMilestone(content []byte, personID string) (*domain.Milestone, error) {
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	var fm MilestoneFrontmatter
	if err := yaml.Unmarshal(frontmatter, &fm); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidFrontmatter, err)
	}

	description, objectives := parseObjectivesFromBody(body)

	return &domain.Milestone{
		ID:          fm.ID,
		PersonID:    personID,
		Number:      fm.Number,
		Name:        fm.Name,
		Description: description,
		Status:      fm.Status,
		Objectives:  objectives,
		CreatedAt:   fm.CreatedAt,
		UpdatedAt:   fm.UpdatedAt,
	}, nil
}

// RenderMilestone renders a milestone to markdown format.
func RenderMilestone(milestone *domain.Milestone) []byte {
	var buf bytes.Buffer

	// Frontmatter
	fm := MilestoneFrontmatter{
		ID:        milestone.ID,
		Number:    milestone.Number,
		Name:      milestone.Name,
		Status:    milestone.Status,
		CreatedAt: milestone.CreatedAt,
		UpdatedAt: milestone.UpdatedAt,
	}
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(fm)
	buf.WriteString("---\n\n")

	// Title
	buf.WriteString(fmt.Sprintf("# Milestone: %s\n\n", milestone.Name))

	// Description
	if milestone.Description != "" {
		buf.WriteString(milestone.Description)
		if !strings.HasSuffix(milestone.Description, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Objectives section
	if len(milestone.Objectives) > 0 {
		buf.WriteString("## Objectives\n\n")

		for _, obj := range milestone.Objectives {
			renderObjective(&buf, &obj)
		}
	}

	return buf.Bytes()
}

// renderObjective renders a single objective to the buffer.
func renderObjective(buf *bytes.Buffer, obj *domain.Objective) {
	// Objective heading
	buf.WriteString(fmt.Sprintf("### %s\n", obj.Name))
	buf.WriteString(fmt.Sprintf("<!-- id: %s -->\n", obj.ID))
	buf.WriteString(fmt.Sprintf("<!-- achieved: %t -->\n\n", obj.Achieved))

	// Description
	if obj.Description != "" {
		buf.WriteString(obj.Description)
		if !strings.HasSuffix(obj.Description, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Review notes
	if obj.ReviewNotes != "" {
		buf.WriteString(fmt.Sprintf("> **Review Notes:** %s\n\n", obj.ReviewNotes))
	}

	// Actions
	if len(obj.Actions) > 0 {
		buf.WriteString("#### Actions\n\n")
		for _, act := range obj.Actions {
			renderAction(buf, &act)
		}
	}
}

// renderAction renders a single action to the buffer.
func renderAction(buf *bytes.Buffer, act *domain.Action) {
	// Impact marker
	marker := "[+]"
	if act.Impact == domain.ImpactNegative {
		marker = "[-]"
	}

	// Format date
	dateStr := act.Date.Format("2006-01-02")

	buf.WriteString(fmt.Sprintf("- **%s %s:** %s\n", marker, dateStr, act.Description))
	buf.WriteString(fmt.Sprintf("  <!-- id: %s -->\n", act.ID))

	if act.Notes != "" {
		// Indent notes
		for _, line := range strings.Split(act.Notes, "\n") {
			buf.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}
	buf.WriteString("\n")
}

// splitFrontmatter splits YAML frontmatter from markdown body.
func splitFrontmatter(content []byte) (frontmatter, body []byte, err error) {
	text := string(content)

	if !strings.HasPrefix(text, "---\n") {
		return nil, nil, domain.ErrInvalidFrontmatter
	}

	// Find closing ---
	rest := text[4:] // Skip opening ---\n
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, nil, domain.ErrInvalidFrontmatter
	}

	frontmatter = []byte(rest[:endIdx])
	body = []byte(strings.TrimPrefix(rest[endIdx+4:], "\n"))

	return frontmatter, body, nil
}

// extractNotesFromBody extracts free-text notes from profile body.
// Skips the first heading (# Name).
func extractNotesFromBody(body []byte) string {
	text := string(body)
	lines := strings.Split(text, "\n")

	var notes []string
	skipFirst := true

	for _, line := range lines {
		if skipFirst && strings.HasPrefix(line, "# ") {
			skipFirst = false
			continue
		}
		if skipFirst && strings.TrimSpace(line) == "" {
			continue
		}
		skipFirst = false
		notes = append(notes, line)
	}

	result := strings.TrimSpace(strings.Join(notes, "\n"))
	return result
}

// parseObjectivesFromBody parses objectives and actions from milestone body.
func parseObjectivesFromBody(body []byte) (description string, objectives []domain.Objective) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Regex patterns
	h1Pattern := regexp.MustCompile(`^# `)
	h2Pattern := regexp.MustCompile(`^## `)
	h3Pattern := regexp.MustCompile(`^### (.+)`)
	h4Pattern := regexp.MustCompile(`^#### Actions`)
	idPattern := regexp.MustCompile(`<!-- id: (.+) -->`)
	achievedPattern := regexp.MustCompile(`<!-- achieved: (true|false) -->`)
	reviewPattern := regexp.MustCompile(`^> \*\*Review Notes:\*\* (.+)`)
	actionPattern := regexp.MustCompile(`^- \*\*\[([+-])\] (\d{4}-\d{2}-\d{2}):\*\* (.+)`)

	var descLines []string
	var currentObj *domain.Objective
	var currentAction *domain.Action
	inObjectives := false
	inActions := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Skip H1 title
		if h1Pattern.MatchString(line) {
			continue
		}

		// Check for ## Objectives section
		if h2Pattern.MatchString(line) {
			if strings.Contains(line, "Objectives") {
				inObjectives = true
			}
			continue
		}

		// Before objectives section = description
		if !inObjectives {
			if strings.TrimSpace(line) != "" || len(descLines) > 0 {
				descLines = append(descLines, line)
			}
			continue
		}

		// Parse objective heading (### Name)
		if matches := h3Pattern.FindStringSubmatch(line); matches != nil {
			// Save previous objective
			if currentObj != nil {
				if currentAction != nil {
					currentObj.Actions = append(currentObj.Actions, *currentAction)
					currentAction = nil
				}
				objectives = append(objectives, *currentObj)
			}
			currentObj = &domain.Objective{
				Name: matches[1],
			}
			inActions = false
			continue
		}

		if currentObj == nil {
			continue
		}

		// Check for #### Actions
		if h4Pattern.MatchString(line) {
			inActions = true
			continue
		}

		// Parse ID comment
		if matches := idPattern.FindStringSubmatch(line); matches != nil {
			if inActions && currentAction != nil {
				currentAction.ID = matches[1]
			} else if !inActions {
				currentObj.ID = matches[1]
			}
			continue
		}

		// Parse achieved comment
		if matches := achievedPattern.FindStringSubmatch(line); matches != nil {
			currentObj.Achieved = matches[1] == "true"
			continue
		}

		// Parse review notes
		if matches := reviewPattern.FindStringSubmatch(line); matches != nil {
			currentObj.ReviewNotes = matches[1]
			continue
		}

		// Parse action line
		if matches := actionPattern.FindStringSubmatch(line); matches != nil {
			// Save previous action
			if currentAction != nil {
				currentObj.Actions = append(currentObj.Actions, *currentAction)
			}

			impact := domain.ImpactPositive
			if matches[1] == "-" {
				impact = domain.ImpactNegative
			}

			date, _ := time.Parse("2006-01-02", matches[2])

			currentAction = &domain.Action{
				Impact:      impact,
				Date:        date,
				Description: matches[3],
			}
			continue
		}

		// Collect action notes (indented lines after action)
		if currentAction != nil && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "  <!--") {
			note := strings.TrimPrefix(line, "  ")
			if currentAction.Notes != "" {
				currentAction.Notes += "\n"
			}
			currentAction.Notes += note
			continue
		}

		// Collect objective description (non-empty lines before Actions)
		if !inActions && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "<!--") && !strings.HasPrefix(line, ">") {
			if currentObj.Description != "" {
				currentObj.Description += "\n"
			}
			currentObj.Description += line
		}
	}

	// Save last objective and action
	if currentObj != nil {
		if currentAction != nil {
			currentObj.Actions = append(currentObj.Actions, *currentAction)
		}
		objectives = append(objectives, *currentObj)
	}

	description = strings.TrimSpace(strings.Join(descLines, "\n"))
	return description, objectives
}
