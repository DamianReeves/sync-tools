package wizard

import (
	"fmt"
)

// StateTransition represents a valid state transition with compile-time safety
type StateTransition[From, To WizardState] struct {
	from From
	to   To
}

// StateMachine provides type-safe state transitions for the wizard
type StateMachine struct {
	currentState WizardState
	history      []WizardState // For navigation back/forward
	historyIndex int
}

// NewStateMachine creates a new state machine starting with InitialState
func NewStateMachine() *StateMachine {
	initialState := InitialState{}
	return &StateMachine{
		currentState: initialState,
		history:      []WizardState{initialState},
		historyIndex: 0,
	}
}

// CurrentState returns the current state with type assertion helpers
func (sm *StateMachine) CurrentState() WizardState {
	return sm.currentState
}

// IsInState checks if the current state matches the given type
func (sm *StateMachine) IsInState(stateType WizardState) bool {
	switch sm.currentState.(type) {
	case InitialState:
		_, ok := stateType.(InitialState)
		return ok
	case SourceSelectionState:
		_, ok := stateType.(SourceSelectionState)
		return ok
	case DestinationSelectionState:
		_, ok := stateType.(DestinationSelectionState)
		return ok
	case SyncOptionsState:
		_, ok := stateType.(SyncOptionsState)
		return ok
	case ExclusionPatternsState:
		_, ok := stateType.(ExclusionPatternsState)
		return ok
	case DirectoryFilterState:
		_, ok := stateType.(DirectoryFilterState)
		return ok
	case ProgressState:
		_, ok := stateType.(ProgressState)
		return ok
	case CompleteState:
		_, ok := stateType.(CompleteState)
		return ok
	}
	return false
}

// TransitionTo safely transitions to a new state with validation
func (sm *StateMachine) TransitionTo(newState WizardState) error {
	if !sm.isValidTransition(sm.currentState, newState) {
		return fmt.Errorf("invalid state transition from %T to %T", sm.currentState, newState)
	}

	// Add to history if it's a forward navigation
	if sm.historyIndex == len(sm.history)-1 {
		sm.history = append(sm.history, newState)
		sm.historyIndex = len(sm.history) - 1
	} else {
		// If we're not at the end of history, truncate and add new state
		sm.history = sm.history[:sm.historyIndex+1]
		sm.history = append(sm.history, newState)
		sm.historyIndex = len(sm.history) - 1
	}

	sm.currentState = newState
	return nil
}

// CanGoBack checks if navigation back is possible
func (sm *StateMachine) CanGoBack() bool {
	return sm.historyIndex > 0
}

// CanGoForward checks if navigation forward is possible
func (sm *StateMachine) CanGoForward() bool {
	return sm.historyIndex < len(sm.history)-1
}

// GoBack navigates to the previous state
func (sm *StateMachine) GoBack() error {
	if !sm.CanGoBack() {
		return fmt.Errorf("cannot go back from current state")
	}

	sm.historyIndex--
	sm.currentState = sm.history[sm.historyIndex]
	return nil
}

// GoForward navigates to the next state in history
func (sm *StateMachine) GoForward() error {
	if !sm.CanGoForward() {
		return fmt.Errorf("cannot go forward from current state")
	}

	sm.historyIndex++
	sm.currentState = sm.history[sm.historyIndex]
	return nil
}

// isValidTransition checks if a state transition is valid
func (sm *StateMachine) isValidTransition(from, to WizardState) bool {
	switch from.(type) {
	case InitialState:
		_, ok := to.(SourceSelectionState)
		return ok

	case SourceSelectionState:
		switch to.(type) {
		case InitialState, DestinationSelectionState:
			return true
		}

	case DestinationSelectionState:
		switch to.(type) {
		case SourceSelectionState, SyncOptionsState:
			return true
		}

	case SyncOptionsState:
		switch to.(type) {
		case DestinationSelectionState, ExclusionPatternsState:
			return true
		}

	case ExclusionPatternsState:
		switch to.(type) {
		case SyncOptionsState, DirectoryFilterState:
			return true
		}

	case DirectoryFilterState:
		switch to.(type) {
		case ExclusionPatternsState, ProgressState:
			return true
		}

	case ProgressState:
		switch to.(type) {
		case DirectoryFilterState, CompleteState:
			return true
		}

	case CompleteState:
		// Complete state is terminal
		return false
	}

	return false
}

// Type-safe state accessor methods
func (sm *StateMachine) AsInitialState() (InitialState, bool) {
	state, ok := sm.currentState.(InitialState)
	return state, ok
}

func (sm *StateMachine) AsSourceSelectionState() (SourceSelectionState, bool) {
	state, ok := sm.currentState.(SourceSelectionState)
	return state, ok
}

func (sm *StateMachine) AsDestinationSelectionState() (DestinationSelectionState, bool) {
	state, ok := sm.currentState.(DestinationSelectionState)
	return state, ok
}

func (sm *StateMachine) AsSyncOptionsState() (SyncOptionsState, bool) {
	state, ok := sm.currentState.(SyncOptionsState)
	return state, ok
}

func (sm *StateMachine) AsExclusionPatternsState() (ExclusionPatternsState, bool) {
	state, ok := sm.currentState.(ExclusionPatternsState)
	return state, ok
}

func (sm *StateMachine) AsDirectoryFilterState() (DirectoryFilterState, bool) {
	state, ok := sm.currentState.(DirectoryFilterState)
	return state, ok
}

func (sm *StateMachine) AsProgressState() (ProgressState, bool) {
	state, ok := sm.currentState.(ProgressState)
	return state, ok
}

func (sm *StateMachine) AsCompleteState() (CompleteState, bool) {
	state, ok := sm.currentState.(CompleteState)
	return state, ok
}

// State-specific operations that are only available in certain states
type InitialStateOperations struct {
	sm *StateMachine
}

type SourceSelectionOperations struct {
	sm    *StateMachine
	state SourceSelectionState
}

type DestinationSelectionOperations struct {
	sm    *StateMachine
	state DestinationSelectionState
}

type SyncOptionsOperations struct {
	sm    *StateMachine
	state SyncOptionsState
}

// GetInitialOperations returns operations available only in InitialState
func (sm *StateMachine) GetInitialOperations() (*InitialStateOperations, error) {
	if _, ok := sm.AsInitialState(); !ok {
		return nil, fmt.Errorf("current state is not InitialState")
	}
	return &InitialStateOperations{sm: sm}, nil
}

// GetSourceSelectionOperations returns operations available only in SourceSelectionState
func (sm *StateMachine) GetSourceSelectionOperations() (*SourceSelectionOperations, error) {
	state, ok := sm.AsSourceSelectionState()
	if !ok {
		return nil, fmt.Errorf("current state is not SourceSelectionState")
	}
	return &SourceSelectionOperations{sm: sm, state: state}, nil
}

// GetDestinationSelectionOperations returns operations available only in DestinationSelectionState
func (sm *StateMachine) GetDestinationSelectionOperations() (*DestinationSelectionOperations, error) {
	state, ok := sm.AsDestinationSelectionState()
	if !ok {
		return nil, fmt.Errorf("current state is not DestinationSelectionState")
	}
	return &DestinationSelectionOperations{sm: sm, state: state}, nil
}

// GetSyncOptionsOperations returns operations available only in SyncOptionsState
func (sm *StateMachine) GetSyncOptionsOperations() (*SyncOptionsOperations, error) {
	state, ok := sm.AsSyncOptionsState()
	if !ok {
		return nil, fmt.Errorf("current state is not SyncOptionsState")
	}
	return &SyncOptionsOperations{sm: sm, state: state}, nil
}

// Type-safe operations for each state
func (ops *InitialStateOperations) StartSourceSelection() error {
	browser := NewDirectoryBrowser(".")
	newState := SourceSelectionState{
		CurrentPath: ".",
		Directories: []DirectoryInfo{},
		Browser:     browser,
	}
	return ops.sm.TransitionTo(newState)
}

func (ops *SourceSelectionOperations) SelectSource(sourcePath string) error {
	browser := NewDirectoryBrowser(".")
	newState := DestinationSelectionState{
		SourcePath:  sourcePath,
		CurrentPath: ".",
		Directories: []DirectoryInfo{},
		Browser:     browser,
	}
	return ops.sm.TransitionTo(newState)
}

func (ops *DestinationSelectionOperations) SelectDestination(destPath string) error {
	newState := SyncOptionsState{
		SourcePath:       ops.state.SourcePath,
		DestinationPath:  destPath,
		Mode:             "one-way",
		DryRun:           false,
		HiddenDirs:       true,
		UseGitIgnore:     false,
		ConflictStrategy: "newest-wins",
		Editor:           nil,
	}
	return ops.sm.TransitionTo(newState)
}

func (ops *SyncOptionsOperations) ConfigureOptions(mode string, dryRun bool, hiddenDirs bool, useGitIgnore bool, conflictStrategy string) error {
	// Update the current state with new options
	updatedState := SyncOptionsState{
		SourcePath:       ops.state.SourcePath,
		DestinationPath:  ops.state.DestinationPath,
		Mode:             mode,
		DryRun:           dryRun,
		HiddenDirs:       hiddenDirs,
		UseGitIgnore:     useGitIgnore,
		ConflictStrategy: conflictStrategy,
		Editor:           ops.state.Editor,
	}
	ops.sm.currentState = updatedState
	return nil
}

func (ops *SyncOptionsOperations) ProceedToExclusionPatterns() error {
	newState := ExclusionPatternsState{
		SourcePath:      ops.state.SourcePath,
		DestinationPath: ops.state.DestinationPath,
		SyncOptions:     ops.state,
		Patterns:        []ExclusionPattern{{Pattern: ".git/", Source: "default", Valid: true}},
	}
	return ops.sm.TransitionTo(newState)
}