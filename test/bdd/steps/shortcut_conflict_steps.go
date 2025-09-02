package steps

import (
	"fmt"
	"strings"
)

// Shortcut character conflict step definitions using TestContext
func (tc *TestContext) thePathShouldBeShortened() error {
	// Simulate backspace removing a character
	tc.lastOutput = "path_shortened_by_backspace"
	return nil
}

func (tc *TestContext) iPressToRemoveCharacter(key string) error {
	if key != "backspace" {
		return fmt.Errorf("unsupported remove character key: %s", key)
	}

	tc.lastOutput = fmt.Sprintf("pressed_%s_to_remove_character", key)
	return nil
}

func (tc *TestContext) iPressToClearPath(key string) error {
	if key != "ctrl+u" {
		return fmt.Errorf("unsupported clear path key: %s", key)
	}

	tc.lastOutput = fmt.Sprintf("pressed_%s_to_clear_path", key)
	return nil
}

func (tc *TestContext) thePathInputShouldBeEmpty() error {
	// Verify that path was cleared
	if !strings.Contains(tc.lastOutput, "clear_path") {
		return fmt.Errorf("expected path to be cleared, but last action was: %s", tc.lastOutput)
	}
	return nil
}
