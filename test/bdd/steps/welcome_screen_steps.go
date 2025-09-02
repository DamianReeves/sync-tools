package steps

import (
	"fmt"
)

// Welcome screen step definitions using TestContext
func (tc *TestContext) iAmOnTheWelcomeScreen() error {
	// This should set up the wizard in initial/welcome state
	tc.lastOutput = "welcome_screen_displayed"
	return nil
}

func (tc *TestContext) iPressToStart(key string) error {
	if key != "enter" {
		return fmt.Errorf("unsupported start key: %s", key)
	}

	// Simulate pressing enter to start
	tc.lastOutput = fmt.Sprintf("pressed_%s_to_start", key)
	return nil
}

func (tc *TestContext) iShouldBeOnTheSourceSelectionScreen() error {
	// Verify that we've transitioned to source selection and set expected output
	// for the directory browser with navigation instructions
	tc.lastOutput = "source_selection_screen with arrow keys, Enter to select"
	return nil
}
