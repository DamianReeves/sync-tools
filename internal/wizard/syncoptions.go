package wizard

import (
	"fmt"
	"strings"
)

// SyncOptionsEditor manages interactive sync options configuration
type SyncOptionsEditor struct {
	options       *SyncOptionsState
	selectedField int
	fields        []SyncOptionField
}

// SyncOptionField represents a configurable sync option
type SyncOptionField struct {
	Name        string
	Type        FieldType
	Value       interface{}
	Options     []string // For select/radio fields
	Description string
}

type FieldType int

const (
	FieldTypeRadio FieldType = iota
	FieldTypeCheckbox
	FieldTypeSelect
	FieldTypeNumber
)

// NewSyncOptionsEditor creates a new sync options editor
func NewSyncOptionsEditor(state *SyncOptionsState) *SyncOptionsEditor {
	editor := &SyncOptionsEditor{
		options:       state,
		selectedField: 0,
	}

	editor.initializeFields()
	return editor
}

// initializeFields sets up the configurable fields
func (e *SyncOptionsEditor) initializeFields() {
	e.fields = []SyncOptionField{
		{
			Name:        "Mode",
			Type:        FieldTypeRadio,
			Value:       e.options.Mode,
			Options:     []string{"one-way", "two-way"},
			Description: "Sync direction mode",
		},
		{
			Name:        "Dry Run",
			Type:        FieldTypeCheckbox,
			Value:       e.options.DryRun,
			Description: "Preview changes without executing",
		},
		{
			Name:        "Hidden Directories",
			Type:        FieldTypeCheckbox,
			Value:       e.options.HiddenDirs,
			Description: "Exclude hidden directories (starting with .)",
		},
		{
			Name:        "Use Git Ignore",
			Type:        FieldTypeCheckbox,
			Value:       e.options.UseGitIgnore,
			Description: "Apply .gitignore patterns from source",
		},
		{
			Name:        "Conflict Strategy",
			Type:        FieldTypeSelect,
			Value:       e.options.ConflictStrategy,
			Options:     []string{"newest-wins", "source-wins", "dest-wins", "backup"},
			Description: "How to resolve file conflicts",
		},
	}
}

// GetSelectedField returns the currently selected field index
func (e *SyncOptionsEditor) GetSelectedField() int {
	return e.selectedField
}

// GetFields returns all configurable fields
func (e *SyncOptionsEditor) GetFields() []SyncOptionField {
	return e.fields
}

// MoveUp moves selection to previous field
func (e *SyncOptionsEditor) MoveUp() {
	if e.selectedField > 0 {
		e.selectedField--
	}
}

// MoveDown moves selection to next field
func (e *SyncOptionsEditor) MoveDown() {
	if e.selectedField < len(e.fields)-1 {
		e.selectedField++
	}
}

// ToggleValue toggles checkbox or cycles through options for current field
func (e *SyncOptionsEditor) ToggleValue() {
	field := &e.fields[e.selectedField]

	switch field.Type {
	case FieldTypeCheckbox:
		field.Value = !field.Value.(bool)
		e.updateOptionsState()

	case FieldTypeRadio, FieldTypeSelect:
		currentValue := field.Value.(string)
		currentIndex := -1

		// Find current value index
		for i, option := range field.Options {
			if option == currentValue {
				currentIndex = i
				break
			}
		}

		// Cycle to next option
		nextIndex := (currentIndex + 1) % len(field.Options)
		field.Value = field.Options[nextIndex]
		e.updateOptionsState()
	}
}

// ChangeValue changes value using left/right arrow keys
func (e *SyncOptionsEditor) ChangeValue(direction int) {
	field := &e.fields[e.selectedField]

	switch field.Type {
	case FieldTypeRadio, FieldTypeSelect:
		currentValue := field.Value.(string)
		currentIndex := -1

		// Find current value index
		for i, option := range field.Options {
			if option == currentValue {
				currentIndex = i
				break
			}
		}

		// Move in specified direction
		newIndex := currentIndex + direction
		if newIndex < 0 {
			newIndex = len(field.Options) - 1
		} else if newIndex >= len(field.Options) {
			newIndex = 0
		}

		field.Value = field.Options[newIndex]
		e.updateOptionsState()
	}
}

// updateOptionsState updates the underlying sync options state
func (e *SyncOptionsEditor) updateOptionsState() {
	for _, field := range e.fields {
		switch field.Name {
		case "Mode":
			e.options.Mode = field.Value.(string)
		case "Dry Run":
			e.options.DryRun = field.Value.(bool)
		case "Hidden Directories":
			e.options.HiddenDirs = field.Value.(bool)
		case "Use Git Ignore":
			e.options.UseGitIgnore = field.Value.(bool)
		case "Conflict Strategy":
			e.options.ConflictStrategy = field.Value.(string)
		}
	}
}

// RenderField renders a single option field
func (e *SyncOptionsEditor) RenderField(fieldIndex int, isSelected bool) string {
	field := e.fields[fieldIndex]

	// Selection indicator
	prefix := "  "
	if isSelected {
		prefix = "▶ "
	}

	// Field value display
	var valueDisplay string
	switch field.Type {
	case FieldTypeCheckbox:
		checked := field.Value.(bool)
		if checked {
			valueDisplay = "[✓]"
		} else {
			valueDisplay = "[ ]"
		}

	case FieldTypeRadio:
		currentValue := field.Value.(string)
		var options []string
		for _, option := range field.Options {
			if option == currentValue {
				options = append(options, fmt.Sprintf("(•) %s", option))
			} else {
				options = append(options, fmt.Sprintf("( ) %s", option))
			}
		}
		valueDisplay = strings.Join(options, " ")

	case FieldTypeSelect:
		valueDisplay = fmt.Sprintf("<%s>", field.Value.(string))

	default:
		valueDisplay = fmt.Sprintf("%v", field.Value)
	}

	return fmt.Sprintf("%s%-20s %s", prefix, field.Name+":", valueDisplay)
}

// RenderAllFields renders all option fields
func (e *SyncOptionsEditor) RenderAllFields() string {
	var content strings.Builder

	for i, field := range e.fields {
		isSelected := i == e.selectedField
		content.WriteString(e.RenderField(i, isSelected))
		content.WriteString("\n")

		// Add description for selected field
		if isSelected {
			content.WriteString(fmt.Sprintf("    %s\n", field.Description))
		}
	}

	return content.String()
}
