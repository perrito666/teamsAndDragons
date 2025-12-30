package tui

import (
	"fmt"
	"strings"
	"time"

	"teamsAndDragons/internal/domain"
	"teamsAndDragons/internal/service"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// View represents the current view in the TUI.
type View int

const (
	ViewPeopleList View = iota
	ViewPersonProfile
	ViewMilestoneDetail
	ViewObjectiveReview
	ViewAddPerson
	ViewEditPerson
	ViewAddMilestone
	ViewEditMilestone
	ViewAddObjective
	ViewEditObjective
	ViewAddAction
	ViewEditAction
	ViewConfirmDelete
)

// Model is the main application model.
type Model struct {
	service service.TrackerService
	styles  Styles

	// Current view
	view     View
	prevView View

	// Window dimensions
	width  int
	height int

	// Data
	people     []domain.Person
	milestones []domain.Milestone

	// Selected items
	selectedPerson    *domain.Person
	selectedMilestone *domain.Milestone
	selectedObjective *domain.Objective
	selectedAction    *domain.Action

	// List components
	peopleList    list.Model
	milestoneList list.Model
	objectiveList list.Model
	actionList    list.Model

	// Input components
	nameInput        textinput.Model
	descriptionInput textarea.Model
	notesInput       textarea.Model
	dateInput        textinput.Model
	impactChoice     int // 0 = positive, 1 = negative

	// Form state
	formMode        string // "add" or "edit"
	deleteType      string // "person", "milestone", "objective", "action"
	goBackAfterSave bool   // Whether to go back to listing after save

	// Error message
	err error

	// Focus index for forms
	focusIndex int
}

// listItem implements list.Item for our domain types.
type listItem struct {
	title       string
	description string
	id          string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

// Key bindings
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Add      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Review   key.Binding
	Toggle   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Save     key.Binding
	Cancel   key.Binding
	Export   key.Binding
	Help     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Review: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "review"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("t", " "),
		key.WithHelp("t/space", "toggle achieved"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Export: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "export HTML"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// NewModel creates a new TUI model.
func NewModel(svc service.TrackerService) Model {
	styles := DefaultStyles()

	// Initialize text inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter name..."
	nameInput.Focus()

	descriptionInput := textarea.New()
	descriptionInput.Placeholder = "Enter description..."
	descriptionInput.SetHeight(3)

	notesInput := textarea.New()
	notesInput.Placeholder = "Enter notes..."
	notesInput.SetHeight(5)

	dateInput := textinput.New()
	dateInput.Placeholder = "YYYY-MM-DD"

	// Initialize lists
	peopleList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	peopleList.Title = "People"
	peopleList.SetShowHelp(false)
	peopleList.SetFilteringEnabled(true)

	milestoneList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	milestoneList.Title = "Milestones"
	milestoneList.SetShowHelp(false)
	milestoneList.SetFilteringEnabled(false)

	objectiveList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	objectiveList.Title = "Objectives"
	objectiveList.SetShowHelp(false)
	objectiveList.SetFilteringEnabled(false)

	actionList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	actionList.Title = "Actions"
	actionList.SetShowHelp(false)
	actionList.SetFilteringEnabled(false)

	return Model{
		service:          svc,
		styles:           styles,
		view:             ViewPeopleList,
		peopleList:       peopleList,
		milestoneList:    milestoneList,
		objectiveList:    objectiveList,
		actionList:       actionList,
		nameInput:        nameInput,
		descriptionInput: descriptionInput,
		notesInput:       notesInput,
		dateInput:        dateInput,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.loadPeople()
}

// loadPeople returns a command to load people.
func (m Model) loadPeople() tea.Cmd {
	return func() tea.Msg {
		people, err := m.service.ListPeople()
		if err != nil {
			return errMsg{err}
		}
		return peopleLoadedMsg{people}
	}
}

// loadMilestones returns a command to load milestones.
func (m Model) loadMilestones(personID string) tea.Cmd {
	return func() tea.Msg {
		milestones, err := m.service.ListMilestones(personID)
		if err != nil {
			return errMsg{err}
		}
		return milestonesLoadedMsg{milestones}
	}
}

// Messages
type errMsg struct{ err error }
type peopleLoadedMsg struct{ people []domain.Person }
type milestonesLoadedMsg struct{ milestones []domain.Milestone }
type personSavedMsg struct{ person *domain.Person }
type milestoneSavedMsg struct{ milestone *domain.Milestone }
type deletedMsg struct{}
type exportedMsg struct{ path string }

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.peopleList.SetSize(msg.Width-4, msg.Height-8)
		m.milestoneList.SetSize(msg.Width-4, msg.Height-10)
		m.objectiveList.SetSize(msg.Width-4, msg.Height-10)
		m.actionList.SetSize(msg.Width-4, msg.Height-10)
		// Set input widths to fill screen
		inputWidth := msg.Width - 8
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.nameInput.Width = inputWidth
		m.dateInput.Width = inputWidth
		m.descriptionInput.SetWidth(inputWidth)
		m.descriptionInput.SetHeight(5)
		m.notesInput.SetWidth(inputWidth)
		m.notesInput.SetHeight(8)
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case peopleLoadedMsg:
		m.people = msg.people
		m.updatePeopleList()
		return m, nil

	case milestonesLoadedMsg:
		m.milestones = msg.milestones
		m.updateMilestoneList()
		return m, nil

	case personSavedMsg:
		m.selectedPerson = msg.person
		// Only go back to profile if goBackAfterSave is set
		if m.goBackAfterSave {
			m.view = ViewPersonProfile
			m.goBackAfterSave = false
		}
		return m, m.loadPeople()

	case milestoneSavedMsg:
		m.selectedMilestone = msg.milestone
		m.updateObjectiveList()
		// Also update selected objective if we have one
		if m.selectedObjective != nil {
			for i := range m.selectedMilestone.Objectives {
				if m.selectedMilestone.Objectives[i].ID == m.selectedObjective.ID {
					m.selectedObjective = &m.selectedMilestone.Objectives[i]
					m.updateActionList()
					break
				}
			}
		}
		// Return to appropriate view based on what was being edited (only if goBackAfterSave is set)
		if m.goBackAfterSave {
			switch m.view {
			case ViewAddMilestone, ViewEditMilestone:
				m.view = ViewPersonProfile
			case ViewAddObjective, ViewEditObjective:
				m.view = ViewMilestoneDetail
			case ViewAddAction, ViewEditAction:
				m.view = ViewObjectiveReview
			}
			m.goBackAfterSave = false
		}
		return m, m.loadMilestones(m.selectedPerson.ID)

	case deletedMsg:
		switch m.deleteType {
		case "person":
			m.selectedPerson = nil
			m.view = ViewPeopleList
			return m, m.loadPeople()
		case "milestone":
			m.selectedMilestone = nil
			m.view = ViewPersonProfile
			return m, m.loadMilestones(m.selectedPerson.ID)
		case "objective":
			m.selectedObjective = nil
			m.view = ViewMilestoneDetail
			return m, m.loadMilestones(m.selectedPerson.ID)
		case "action":
			m.selectedAction = nil
			m.view = ViewObjectiveReview
			return m, m.loadMilestones(m.selectedPerson.ID)
		}
		return m, nil

	case exportedMsg:
		// Show success message
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Update active list
	switch m.view {
	case ViewPeopleList:
		var cmd tea.Cmd
		m.peopleList, cmd = m.peopleList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewPersonProfile:
		var cmd tea.Cmd
		m.milestoneList, cmd = m.milestoneList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewMilestoneDetail:
		var cmd tea.Cmd
		m.objectiveList, cmd = m.objectiveList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewObjectiveReview:
		var cmd tea.Cmd
		m.actionList, cmd = m.actionList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewAddPerson, ViewEditPerson, ViewAddMilestone, ViewEditMilestone,
		ViewAddObjective, ViewEditObjective, ViewAddAction, ViewEditAction:
		cmds = append(cmds, m.updateFormInputs(msg)...)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if key.Matches(msg, keys.Quit) && !m.isFormView() {
		return m, tea.Quit
	}

	switch m.view {
	case ViewPeopleList:
		return m.handlePeopleListKeys(msg)
	case ViewPersonProfile:
		return m.handlePersonProfileKeys(msg)
	case ViewMilestoneDetail:
		return m.handleMilestoneDetailKeys(msg)
	case ViewObjectiveReview:
		return m.handleObjectiveReviewKeys(msg)
	case ViewAddPerson, ViewEditPerson:
		return m.handlePersonFormKeys(msg)
	case ViewAddMilestone, ViewEditMilestone:
		return m.handleMilestoneFormKeys(msg)
	case ViewAddObjective, ViewEditObjective:
		return m.handleObjectiveFormKeys(msg)
	case ViewAddAction, ViewEditAction:
		return m.handleActionFormKeys(msg)
	case ViewConfirmDelete:
		return m.handleDeleteConfirmKeys(msg)
	}

	return m, nil
}

func (m *Model) handlePeopleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if item, ok := m.peopleList.SelectedItem().(listItem); ok {
			person, _ := m.service.GetPerson(item.id)
			if person != nil {
				m.selectedPerson = person
				m.view = ViewPersonProfile
				return m, m.loadMilestones(person.ID)
			}
		}
	case key.Matches(msg, keys.Add):
		m.view = ViewAddPerson
		m.formMode = "add"
		m.resetForm()
		m.nameInput.Focus()
		return m, nil
	case key.Matches(msg, keys.Edit):
		if item, ok := m.peopleList.SelectedItem().(listItem); ok {
			person, _ := m.service.GetPerson(item.id)
			if person != nil {
				m.selectedPerson = person
				m.view = ViewEditPerson
				m.formMode = "edit"
				m.nameInput.SetValue(person.Name)
				m.notesInput.SetValue(person.Notes)
				m.nameInput.Focus()
			}
		}
		return m, nil
	case key.Matches(msg, keys.Delete):
		if item, ok := m.peopleList.SelectedItem().(listItem); ok {
			person, _ := m.service.GetPerson(item.id)
			if person != nil {
				m.selectedPerson = person
				m.prevView = m.view
				m.view = ViewConfirmDelete
				m.deleteType = "person"
			}
		}
		return m, nil
	}
	// Pass navigation keys to list
	var cmd tea.Cmd
	m.peopleList, cmd = m.peopleList.Update(msg)
	return m, cmd
}

func (m *Model) handlePersonProfileKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.view = ViewPeopleList
		m.selectedPerson = nil
		return m, nil
	case key.Matches(msg, keys.Enter):
		if item, ok := m.milestoneList.SelectedItem().(listItem); ok {
			milestone, _ := m.service.GetMilestone(item.id)
			if milestone != nil {
				m.selectedMilestone = milestone
				m.view = ViewMilestoneDetail
				m.updateObjectiveList()
			}
		}
		return m, nil
	case key.Matches(msg, keys.Add):
		m.view = ViewAddMilestone
		m.formMode = "add"
		m.resetForm()
		m.nameInput.Focus()
		return m, nil
	case key.Matches(msg, keys.Edit):
		if m.selectedPerson != nil {
			m.view = ViewEditPerson
			m.formMode = "edit"
			m.nameInput.SetValue(m.selectedPerson.Name)
			m.notesInput.SetValue(m.selectedPerson.Notes)
			m.nameInput.Focus()
		}
		return m, nil
	case key.Matches(msg, keys.Delete):
		if item, ok := m.milestoneList.SelectedItem().(listItem); ok {
			milestone, _ := m.service.GetMilestone(item.id)
			if milestone != nil {
				m.selectedMilestone = milestone
				m.prevView = m.view
				m.view = ViewConfirmDelete
				m.deleteType = "milestone"
			}
		}
		return m, nil
	case key.Matches(msg, keys.Export):
		if m.selectedPerson != nil {
			return m, func() tea.Msg {
				path := m.selectedPerson.Slug() + "-export.html"
				err := m.service.ExportPerson(m.selectedPerson.ID, path, true)
				if err != nil {
					return errMsg{err}
				}
				return exportedMsg{path}
			}
		}
		return m, nil
	}
	// Pass navigation keys to list
	var cmd tea.Cmd
	m.milestoneList, cmd = m.milestoneList.Update(msg)
	return m, cmd
}

func (m *Model) handleMilestoneDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.view = ViewPersonProfile
		m.selectedMilestone = nil
		return m, nil
	case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Review):
		if item, ok := m.objectiveList.SelectedItem().(listItem); ok {
			for i := range m.selectedMilestone.Objectives {
				if m.selectedMilestone.Objectives[i].ID == item.id {
					m.selectedObjective = &m.selectedMilestone.Objectives[i]
					m.view = ViewObjectiveReview
					m.updateActionList()
					break
				}
			}
		}
		return m, nil
	case key.Matches(msg, keys.Add):
		m.view = ViewAddObjective
		m.formMode = "add"
		m.resetForm()
		m.nameInput.Focus()
		return m, nil
	case key.Matches(msg, keys.Edit):
		if item, ok := m.objectiveList.SelectedItem().(listItem); ok {
			for i := range m.selectedMilestone.Objectives {
				if m.selectedMilestone.Objectives[i].ID == item.id {
					obj := &m.selectedMilestone.Objectives[i]
					m.selectedObjective = obj
					m.view = ViewEditObjective
					m.formMode = "edit"
					m.nameInput.SetValue(obj.Name)
					m.descriptionInput.SetValue(obj.Description)
					m.nameInput.Focus()
					break
				}
			}
		}
		return m, nil
	case key.Matches(msg, keys.Toggle):
		if item, ok := m.objectiveList.SelectedItem().(listItem); ok {
			for i := range m.selectedMilestone.Objectives {
				if m.selectedMilestone.Objectives[i].ID == item.id {
					obj := &m.selectedMilestone.Objectives[i]
					return m, func() tea.Msg {
						milestone, err := m.service.SetObjectiveAchieved(
							m.selectedMilestone.ID, obj.ID, !obj.Achieved, "")
						if err != nil {
							return errMsg{err}
						}
						return milestoneSavedMsg{milestone}
					}
				}
			}
		}
		return m, nil
	case key.Matches(msg, keys.Delete):
		if item, ok := m.objectiveList.SelectedItem().(listItem); ok {
			for i := range m.selectedMilestone.Objectives {
				if m.selectedMilestone.Objectives[i].ID == item.id {
					m.selectedObjective = &m.selectedMilestone.Objectives[i]
					m.prevView = m.view
					m.view = ViewConfirmDelete
					m.deleteType = "objective"
					break
				}
			}
		}
		return m, nil
	}
	// Pass navigation keys to list
	var cmd tea.Cmd
	m.objectiveList, cmd = m.objectiveList.Update(msg)
	return m, cmd
}

func (m *Model) handleObjectiveReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.view = ViewMilestoneDetail
		m.selectedObjective = nil
		return m, nil
	case key.Matches(msg, keys.Add):
		m.view = ViewAddAction
		m.formMode = "add"
		m.resetForm()
		m.descriptionInput.Focus()
		m.focusIndex = 0
		return m, nil
	case key.Matches(msg, keys.Edit):
		if item, ok := m.actionList.SelectedItem().(listItem); ok {
			for i := range m.selectedObjective.Actions {
				if m.selectedObjective.Actions[i].ID == item.id {
					act := &m.selectedObjective.Actions[i]
					m.selectedAction = act
					m.view = ViewEditAction
					m.formMode = "edit"
					m.descriptionInput.SetValue(act.Description)
					m.dateInput.SetValue(act.Date.Format("2006-01-02"))
					m.notesInput.SetValue(act.Notes)
					if act.Impact == domain.ImpactPositive {
						m.impactChoice = 0
					} else {
						m.impactChoice = 1
					}
					m.descriptionInput.Focus()
					m.focusIndex = 0
					break
				}
			}
		}
		return m, nil
	case key.Matches(msg, keys.Toggle):
		return m, func() tea.Msg {
			milestone, err := m.service.SetObjectiveAchieved(
				m.selectedMilestone.ID, m.selectedObjective.ID,
				!m.selectedObjective.Achieved, "")
			if err != nil {
				return errMsg{err}
			}
			return milestoneSavedMsg{milestone}
		}
	case key.Matches(msg, keys.Delete):
		if item, ok := m.actionList.SelectedItem().(listItem); ok {
			for i := range m.selectedObjective.Actions {
				if m.selectedObjective.Actions[i].ID == item.id {
					m.selectedAction = &m.selectedObjective.Actions[i]
					m.prevView = m.view
					m.view = ViewConfirmDelete
					m.deleteType = "action"
					break
				}
			}
		}
		return m, nil
	}
	// Pass navigation keys to list
	var cmd tea.Cmd
	m.actionList, cmd = m.actionList.Update(msg)
	return m, cmd
}

func (m *Model) handlePersonFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		if m.formMode == "add" {
			m.view = ViewPeopleList
		} else {
			m.view = ViewPersonProfile
		}
		return m, nil
	case key.Matches(msg, keys.Tab):
		m.focusIndex = (m.focusIndex + 1) % 2
		m.updateFormFocus()
		return m, nil
	case key.Matches(msg, keys.Save), msg.String() == "ctrl+s", msg.String() == "S":
		// Capital S means save and go back
		if msg.String() == "S" {
			m.goBackAfterSave = true
		}
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			m.err = fmt.Errorf("name is required")
			return m, nil
		}
		notes := m.notesInput.Value()

		if m.formMode == "add" {
			m.goBackAfterSave = true // Always go back after add
			return m, func() tea.Msg {
				person, err := m.service.CreatePerson(name, notes)
				if err != nil {
					return errMsg{err}
				}
				return personSavedMsg{person}
			}
		}
		return m, func() tea.Msg {
			person, err := m.service.UpdatePerson(m.selectedPerson.ID, name, notes)
			if err != nil {
				return errMsg{err}
			}
			return personSavedMsg{person}
		}
	}
	// Pass other keys to input components
	return m, tea.Batch(m.updateFormInputs(msg)...)
}

func (m *Model) handleMilestoneFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.view = ViewPersonProfile
		return m, nil
	case key.Matches(msg, keys.Tab):
		m.focusIndex = (m.focusIndex + 1) % 2
		m.updateFormFocus()
		return m, nil
	case key.Matches(msg, keys.Save), msg.String() == "ctrl+s", msg.String() == "S":
		// Capital S means save and go back
		if msg.String() == "S" {
			m.goBackAfterSave = true
		}
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			m.err = fmt.Errorf("name is required")
			return m, nil
		}
		desc := m.descriptionInput.Value()

		if m.formMode == "add" {
			m.goBackAfterSave = true // Always go back after add
			return m, func() tea.Msg {
				milestone, err := m.service.CreateMilestone(m.selectedPerson.ID, name, desc)
				if err != nil {
					return errMsg{err}
				}
				return milestoneSavedMsg{milestone}
			}
		}
		return m, func() tea.Msg {
			milestone, err := m.service.UpdateMilestone(
				m.selectedMilestone.ID, name, desc, m.selectedMilestone.Status)
			if err != nil {
				return errMsg{err}
			}
			return milestoneSavedMsg{milestone}
		}
	}
	// Pass other keys to input components
	return m, tea.Batch(m.updateFormInputs(msg)...)
}

func (m *Model) handleObjectiveFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.view = ViewMilestoneDetail
		return m, nil
	case key.Matches(msg, keys.Tab):
		m.focusIndex = (m.focusIndex + 1) % 2
		m.updateFormFocus()
		return m, nil
	case key.Matches(msg, keys.Save), msg.String() == "ctrl+s", msg.String() == "S":
		// Capital S means save and go back
		if msg.String() == "S" {
			m.goBackAfterSave = true
		}
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			m.err = fmt.Errorf("name is required")
			return m, nil
		}
		desc := m.descriptionInput.Value()

		if m.formMode == "add" {
			m.goBackAfterSave = true // Always go back after add
			return m, func() tea.Msg {
				milestone, err := m.service.AddObjective(m.selectedMilestone.ID, name, desc)
				if err != nil {
					return errMsg{err}
				}
				return milestoneSavedMsg{milestone}
			}
		}
		return m, func() tea.Msg {
			milestone, err := m.service.UpdateObjective(
				m.selectedMilestone.ID, m.selectedObjective.ID, name, desc)
			if err != nil {
				return errMsg{err}
			}
			return milestoneSavedMsg{milestone}
		}
	}
	// Pass other keys to input components
	return m, tea.Batch(m.updateFormInputs(msg)...)
}

func (m *Model) handleActionFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.view = ViewObjectiveReview
		return m, nil
	case key.Matches(msg, keys.Tab):
		m.focusIndex = (m.focusIndex + 1) % 4
		m.updateActionFormFocus()
		return m, nil
	case msg.String() == "left", msg.String() == "right":
		if m.focusIndex == 2 { // Impact choice
			m.impactChoice = 1 - m.impactChoice
		}
		return m, nil
	case msg.String() == "n" && m.focusIndex == 1: // Date field: 'n' for now/today
		m.dateInput.SetValue(time.Now().Format("2006-01-02"))
		return m, nil
	case msg.String() == "up" && m.focusIndex == 1: // Date field: up arrow adds one day
		m.adjustDate(1)
		return m, nil
	case msg.String() == "down" && m.focusIndex == 1: // Date field: down arrow subtracts one day
		m.adjustDate(-1)
		return m, nil
	case key.Matches(msg, keys.Save), msg.String() == "ctrl+s", msg.String() == "S":
		// Capital S means save and go back
		if msg.String() == "S" {
			m.goBackAfterSave = true
		}
		desc := strings.TrimSpace(m.descriptionInput.Value())
		if desc == "" {
			m.err = fmt.Errorf("description is required")
			return m, nil
		}

		dateStr := m.dateInput.Value()
		date, err := parseDate(dateStr)
		if err != nil {
			m.err = fmt.Errorf("invalid date format (use YYYY-MM-DD)")
			return m, nil
		}

		impact := domain.ImpactPositive
		if m.impactChoice == 1 {
			impact = domain.ImpactNegative
		}
		notes := m.notesInput.Value()

		if m.formMode == "add" {
			m.goBackAfterSave = true // Always go back after add
			return m, func() tea.Msg {
				milestone, err := m.service.AddAction(
					m.selectedMilestone.ID, m.selectedObjective.ID,
					desc, impact, date, notes)
				if err != nil {
					return errMsg{err}
				}
				return milestoneSavedMsg{milestone}
			}
		}
		return m, func() tea.Msg {
			milestone, err := m.service.UpdateAction(
				m.selectedMilestone.ID, m.selectedObjective.ID, m.selectedAction.ID,
				desc, impact, date, notes)
			if err != nil {
				return errMsg{err}
			}
			return milestoneSavedMsg{milestone}
		}
	}
	// Pass other keys to input components
	return m, tea.Batch(m.updateFormInputs(msg)...)
}

func (m *Model) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		switch m.deleteType {
		case "person":
			return m, func() tea.Msg {
				err := m.service.DeletePerson(m.selectedPerson.ID)
				if err != nil {
					return errMsg{err}
				}
				return deletedMsg{}
			}
		case "milestone":
			return m, func() tea.Msg {
				err := m.service.DeleteMilestone(m.selectedMilestone.ID)
				if err != nil {
					return errMsg{err}
				}
				return deletedMsg{}
			}
		case "objective":
			return m, func() tea.Msg {
				_, err := m.service.DeleteObjective(m.selectedMilestone.ID, m.selectedObjective.ID)
				if err != nil {
					return errMsg{err}
				}
				return deletedMsg{}
			}
		case "action":
			return m, func() tea.Msg {
				_, err := m.service.DeleteAction(
					m.selectedMilestone.ID, m.selectedObjective.ID, m.selectedAction.ID)
				if err != nil {
					return errMsg{err}
				}
				return deletedMsg{}
			}
		}
	case "n", "N", "esc":
		m.view = m.prevView
		return m, nil
	}
	return m, nil
}

func (m *Model) updateFormInputs(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.nameInput, cmd = m.nameInput.Update(msg)
	cmds = append(cmds, cmd)

	m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	cmds = append(cmds, cmd)

	m.notesInput, cmd = m.notesInput.Update(msg)
	cmds = append(cmds, cmd)

	m.dateInput, cmd = m.dateInput.Update(msg)
	cmds = append(cmds, cmd)

	return cmds
}

func (m *Model) updateFormFocus() {
	m.nameInput.Blur()
	m.descriptionInput.Blur()
	m.notesInput.Blur()

	switch m.focusIndex {
	case 0:
		m.nameInput.Focus()
	case 1:
		if m.view == ViewAddPerson || m.view == ViewEditPerson {
			m.notesInput.Focus()
		} else {
			m.descriptionInput.Focus()
		}
	}
}

func (m *Model) updateActionFormFocus() {
	m.descriptionInput.Blur()
	m.dateInput.Blur()
	m.notesInput.Blur()

	switch m.focusIndex {
	case 0:
		m.descriptionInput.Focus()
	case 1:
		m.dateInput.Focus()
	case 2:
		// Impact choice - no focus needed
	case 3:
		m.notesInput.Focus()
	}
}

func (m *Model) resetForm() {
	m.nameInput.SetValue("")
	m.descriptionInput.SetValue("")
	m.notesInput.SetValue("")
	m.dateInput.SetValue("")
	m.impactChoice = 0
	m.focusIndex = 0
	m.err = nil
}

func (m *Model) isFormView() bool {
	switch m.view {
	case ViewAddPerson, ViewEditPerson, ViewAddMilestone, ViewEditMilestone,
		ViewAddObjective, ViewEditObjective, ViewAddAction, ViewEditAction,
		ViewConfirmDelete:
		return true
	}
	return false
}

func (m *Model) updatePeopleList() {
	items := make([]list.Item, len(m.people))
	for i, p := range m.people {
		items[i] = listItem{
			title:       p.Name,
			description: truncate(p.Notes, 50),
			id:          p.ID,
		}
	}
	m.peopleList.SetItems(items)
}

func (m *Model) updateMilestoneList() {
	items := make([]list.Item, len(m.milestones))
	for i, ms := range m.milestones {
		status := "In Progress"
		if ms.Status == domain.MilestoneCompleted {
			status = "Completed"
		}
		items[i] = listItem{
			title:       fmt.Sprintf("%d. %s", ms.Number, ms.Name),
			description: fmt.Sprintf("%s | %d objectives | %d%% complete", status, len(ms.Objectives), ms.Progress()),
			id:          ms.ID,
		}
	}
	m.milestoneList.SetItems(items)
}

func (m *Model) updateObjectiveList() {
	if m.selectedMilestone == nil {
		return
	}
	items := make([]list.Item, len(m.selectedMilestone.Objectives))
	for i, obj := range m.selectedMilestone.Objectives {
		status := "Pending"
		if obj.Achieved {
			status = "Achieved"
		}
		pos, neg := obj.Tally()
		items[i] = listItem{
			title:       obj.Name,
			description: fmt.Sprintf("%s | +%d/-%d actions", status, pos, neg),
			id:          obj.ID,
		}
	}
	m.objectiveList.SetItems(items)
}

func (m *Model) updateActionList() {
	if m.selectedObjective == nil {
		return
	}
	items := make([]list.Item, len(m.selectedObjective.Actions))
	for i, act := range m.selectedObjective.Actions {
		impact := "[+]"
		if act.Impact == domain.ImpactNegative {
			impact = "[-]"
		}
		items[i] = listItem{
			title:       fmt.Sprintf("%s %s", impact, act.Description),
			description: fmt.Sprintf("%s | %s", act.Date.Format("2006-01-02"), truncate(act.Notes, 40)),
			id:          act.ID,
		}
	}
	m.actionList.SetItems(items)
}

// View renders the UI.
func (m Model) View() string {
	var s strings.Builder

	// Header with breadcrumb
	s.WriteString(m.renderBreadcrumb())
	s.WriteString("\n\n")

	// Main content
	switch m.view {
	case ViewPeopleList:
		s.WriteString(m.renderPeopleList())
	case ViewPersonProfile:
		s.WriteString(m.renderPersonProfile())
	case ViewMilestoneDetail:
		s.WriteString(m.renderMilestoneDetail())
	case ViewObjectiveReview:
		s.WriteString(m.renderObjectiveReview())
	case ViewAddPerson, ViewEditPerson:
		s.WriteString(m.renderPersonForm())
	case ViewAddMilestone, ViewEditMilestone:
		s.WriteString(m.renderMilestoneForm())
	case ViewAddObjective, ViewEditObjective:
		s.WriteString(m.renderObjectiveForm())
	case ViewAddAction, ViewEditAction:
		s.WriteString(m.renderActionForm())
	case ViewConfirmDelete:
		s.WriteString(m.renderDeleteConfirm())
	}

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	// Footer with help
	s.WriteString("\n\n")
	s.WriteString(m.renderHelp())

	return m.styles.App.Render(s.String())
}

func (m Model) renderBreadcrumb() string {
	parts := []string{"Career Tracker"}

	if m.selectedPerson != nil {
		parts = append(parts, m.selectedPerson.Name)
	}
	if m.selectedMilestone != nil {
		parts = append(parts, m.selectedMilestone.Name)
	}
	if m.selectedObjective != nil {
		parts = append(parts, m.selectedObjective.Name)
	}

	var styled []string
	for _, p := range parts {
		styled = append(styled, m.styles.Breadcrumb.Render(p))
	}

	return m.styles.Title.Render(strings.Join(styled, m.styles.BreadcrumbSep.String()))
}

func (m Model) renderPeopleList() string {
	if len(m.people) == 0 {
		return m.styles.Muted.Render("No people tracked yet. Press 'a' to add someone.")
	}
	return m.peopleList.View()
}

func (m Model) renderPersonProfile() string {
	if m.selectedPerson == nil {
		return ""
	}

	var s strings.Builder

	// Profile info
	s.WriteString(m.styles.CardTitle.Render(m.selectedPerson.Name))
	s.WriteString("\n")

	if m.selectedPerson.Notes != "" {
		s.WriteString(m.styles.Muted.Render(m.selectedPerson.Notes))
		s.WriteString("\n\n")
	}

	// Milestones
	s.WriteString(m.styles.Subtitle.Render("Milestones"))
	s.WriteString("\n")

	if len(m.milestones) == 0 {
		s.WriteString(m.styles.Muted.Render("No milestones yet. Press 'a' to add one."))
	} else {
		s.WriteString(m.milestoneList.View())
	}

	return s.String()
}

func (m Model) renderMilestoneDetail() string {
	if m.selectedMilestone == nil {
		return ""
	}

	var s strings.Builder
	ms := m.selectedMilestone

	// Status badge
	statusStyle := m.styles.StatusInProgress
	if ms.Status == domain.MilestoneCompleted {
		statusStyle = m.styles.StatusCompleted
	}
	s.WriteString(statusStyle.Render(string(ms.Status)))
	s.WriteString("\n\n")

	// Description
	if ms.Description != "" {
		s.WriteString(m.styles.Muted.Render(ms.Description))
		s.WriteString("\n\n")
	}

	// Progress
	progress := ms.Progress()
	s.WriteString(fmt.Sprintf("Progress: %d%% ", progress))
	s.WriteString(m.styles.RenderProgressBar(progress, 20))
	s.WriteString(fmt.Sprintf(" (%d/%d objectives)\n\n", len(ms.AchievedObjectives()), len(ms.Objectives)))

	// Objectives
	s.WriteString(m.styles.Subtitle.Render("Objectives"))
	s.WriteString("\n")

	if len(ms.Objectives) == 0 {
		s.WriteString(m.styles.Muted.Render("No objectives yet. Press 'a' to add one."))
	} else {
		s.WriteString(m.objectiveList.View())
	}

	return s.String()
}

func (m Model) renderObjectiveReview() string {
	if m.selectedObjective == nil {
		return ""
	}

	var s strings.Builder
	obj := m.selectedObjective

	// Status
	statusStyle := m.styles.StatusPending
	statusText := "Pending"
	if obj.Achieved {
		statusStyle = m.styles.StatusAchieved
		statusText = "Achieved"
	}
	s.WriteString(statusStyle.Render(statusText))
	s.WriteString("\n\n")

	// Description
	if obj.Description != "" {
		s.WriteString(m.styles.Muted.Render(obj.Description))
		s.WriteString("\n\n")
	}

	// Tally
	pos, neg := obj.Tally()
	net := pos - neg
	s.WriteString(m.styles.Label.Render("Tally: "))
	s.WriteString(m.styles.TallyPositive.Render(fmt.Sprintf("+%d", pos)))
	s.WriteString(" / ")
	s.WriteString(m.styles.TallyNegative.Render(fmt.Sprintf("-%d", neg)))
	s.WriteString(" = ")
	netStyle := m.styles.TallyNet
	if net > 0 {
		netStyle = m.styles.TallyPositive
	} else if net < 0 {
		netStyle = m.styles.TallyNegative
	}
	s.WriteString(netStyle.Render(fmt.Sprintf("%+d net", net)))
	s.WriteString("\n\n")

	// Review notes
	if obj.ReviewNotes != "" {
		s.WriteString(m.styles.Warning.Render("Review Notes: "))
		s.WriteString(obj.ReviewNotes)
		s.WriteString("\n\n")
	}

	// Actions
	s.WriteString(m.styles.Subtitle.Render("Actions"))
	s.WriteString("\n")

	if len(obj.Actions) == 0 {
		s.WriteString(m.styles.Muted.Render("No actions recorded yet. Press 'a' to add one."))
	} else {
		s.WriteString(m.actionList.View())
	}

	return s.String()
}

func (m Model) renderPersonForm() string {
	title := "Add Person"
	if m.formMode == "edit" {
		title = "Edit Person"
	}

	var s strings.Builder
	s.WriteString(m.styles.CardTitle.Render(title))
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Name"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Notes"))
	s.WriteString("\n")
	s.WriteString(m.notesInput.View())

	return s.String()
}

func (m Model) renderMilestoneForm() string {
	title := "Add Milestone"
	if m.formMode == "edit" {
		title = "Edit Milestone"
	}

	var s strings.Builder
	s.WriteString(m.styles.CardTitle.Render(title))
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Name (e.g., Senior 1 → Senior 2)"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Description"))
	s.WriteString("\n")
	s.WriteString(m.descriptionInput.View())

	return s.String()
}

func (m Model) renderObjectiveForm() string {
	title := "Add Objective"
	if m.formMode == "edit" {
		title = "Edit Objective"
	}

	var s strings.Builder
	s.WriteString(m.styles.CardTitle.Render(title))
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Name"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Description"))
	s.WriteString("\n")
	s.WriteString(m.descriptionInput.View())

	return s.String()
}

func (m Model) renderActionForm() string {
	title := "Add Action"
	if m.formMode == "edit" {
		title = "Edit Action"
	}

	var s strings.Builder
	s.WriteString(m.styles.CardTitle.Render(title))
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Description"))
	s.WriteString("\n")
	s.WriteString(m.descriptionInput.View())
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Date (YYYY-MM-DD)"))
	s.WriteString("\n")
	s.WriteString(m.dateInput.View())
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Impact"))
	s.WriteString("\n")

	positiveStyle := m.styles.Muted
	negativeStyle := m.styles.Muted
	if m.impactChoice == 0 {
		positiveStyle = m.styles.ImpactPositive
	} else {
		negativeStyle = m.styles.ImpactNegative
	}

	s.WriteString(positiveStyle.Render("[+] Positive"))
	s.WriteString("  ")
	s.WriteString(negativeStyle.Render("[-] Negative"))
	s.WriteString("\n")
	s.WriteString(m.styles.Help.Render("(Use ←/→ to change)"))
	s.WriteString("\n\n")

	s.WriteString(m.styles.Label.Render("Notes (optional)"))
	s.WriteString("\n")
	s.WriteString(m.notesInput.View())

	return s.String()
}

func (m Model) renderDeleteConfirm() string {
	var s strings.Builder

	s.WriteString(m.styles.Error.Render("Confirm Delete"))
	s.WriteString("\n\n")

	var name string
	switch m.deleteType {
	case "person":
		name = m.selectedPerson.Name
	case "milestone":
		name = m.selectedMilestone.Name
	case "objective":
		name = m.selectedObjective.Name
	case "action":
		name = m.selectedAction.Description
	}

	s.WriteString(fmt.Sprintf("Are you sure you want to delete %s '%s'?\n\n",
		m.deleteType, name))
	s.WriteString(m.styles.Warning.Render("This action cannot be undone."))
	s.WriteString("\n\n")
	s.WriteString("Press ")
	s.WriteString(m.styles.Value.Render("y"))
	s.WriteString(" to confirm or ")
	s.WriteString(m.styles.Value.Render("n"))
	s.WriteString(" to cancel")

	return s.String()
}

func (m Model) renderHelp() string {
	var help string

	switch m.view {
	case ViewPeopleList:
		help = "↑/↓ navigate • enter select • a add • e edit • d delete • q quit"
	case ViewPersonProfile:
		help = "↑/↓ navigate • enter view milestone • a add • e edit person • d delete • x export • esc back"
	case ViewMilestoneDetail:
		help = "↑/↓ navigate • enter/r review • a add • e edit • t toggle achieved • d delete • esc back"
	case ViewObjectiveReview:
		help = "↑/↓ navigate • a add action • e edit • t toggle achieved • d delete • esc back"
	case ViewAddPerson, ViewEditPerson, ViewAddMilestone, ViewEditMilestone,
		ViewAddObjective, ViewEditObjective:
		help = "tab next field • ctrl+s save • S save & back • esc cancel"
	case ViewAddAction, ViewEditAction:
		help = "tab next field • ←/→ impact • n today • ↑/↓ +/-1 day • ctrl+s save • S save & back • esc cancel"
	case ViewConfirmDelete:
		help = "y confirm • n cancel"
	}

	return m.styles.Help.Render(help)
}

// Helper functions
func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	return time.Parse("2006-01-02", s)
}

// adjustDate adjusts the date input by the given number of days.
func (m *Model) adjustDate(days int) {
	dateStr := m.dateInput.Value()
	if dateStr == "" {
		// If empty, start from today
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// If invalid, start from today
		date = time.Now()
	}

	newDate := date.AddDate(0, 0, days)
	m.dateInput.SetValue(newDate.Format("2006-01-02"))
}
