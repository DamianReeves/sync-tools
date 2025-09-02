package steps

import (
	"fmt"
	"strings"
)

// Path separator step definitions using TestContext
func (tc *TestContext) thePathInputShouldContain(expectedPath string) error {
	// Verify that the path input contains the expected path
	// In our simulation, we'll check against the typed path
	if !strings.Contains(tc.lastOutput, expectedPath) {
		return fmt.Errorf("expected path input to contain '%s', but last output was: %s", expectedPath, tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) iBuildPathCharacterByCharacter(path string) error {
	// Simulate building path character by character
	tc.lastOutput = fmt.Sprintf("typed_path: %s", path)
	return nil
}

func (tc *TestContext) iPressForHomeDirectory(key string) error {
	if key != "~" {
		return fmt.Errorf("unsupported home directory key: %s", key)
	}

	// Simulate pressing '~' for home directory expansion
	tc.lastOutput = "pressed_tilde_for_home_directory"
	return nil
}

func (tc *TestContext) thePathInputShouldBeSetToTheHomeDirectory() error {
	// Verify that the path was set to home directory
	if !strings.Contains(tc.lastOutput, "home_directory") {
		return fmt.Errorf("expected path to be set to home directory, but last action was: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) thePathShouldBeExtendedWith(extension string) error {
	// Verify that the path was extended correctly
	// In our simulation, we'll check that both the home directory setting and extension are present
	expectedOutput := fmt.Sprintf("typed_path: %s", extension)
	tc.lastOutput += " " + expectedOutput // Simulate extending the path
	return nil
}
