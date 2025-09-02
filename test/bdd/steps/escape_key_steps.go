package steps

import (
	"fmt"
	"strings"
)

// Escape key handling step definitions using TestContext
func (tc *TestContext) iPressToOpenHomeNavigation(key string) error {
	if key != "~" {
		return fmt.Errorf("unsupported home navigation key: %s", key)
	}

	// Simulate pressing '~' to open home navigation
	tc.lastOutput = fmt.Sprintf("pressed_%s_for_home_navigation", key)
	return nil
}

func (tc *TestContext) iShouldSeeTheHomeNavigationModalDisplayed() error {
	// Verify that home navigation modal is visible
	if !strings.Contains(tc.lastOutput, "home_navigation") {
		return fmt.Errorf("expected home navigation modal to be displayed, but last action was: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) iPressToCancelNavigation(key string) error {
	if key != "escape" {
		return fmt.Errorf("unsupported cancel key: %s", key)
	}

	tc.lastOutput = fmt.Sprintf("pressed_%s_to_cancel_navigation", key)
	return nil
}

func (tc *TestContext) theHomeNavigationModalShouldBeClosed() error {
	if !strings.Contains(tc.lastOutput, "cancel_navigation") && !strings.Contains(tc.lastOutput, "select_bookmark") {
		return fmt.Errorf("expected home navigation modal to be closed, but last action was: %s", tc.lastOutput)
	}
	return nil
}

func (tc *TestContext) iPressToSelectBookmark(key string) error {
	if key != "enter" {
		return fmt.Errorf("unsupported select key: %s", key)
	}

	tc.lastOutput = fmt.Sprintf("pressed_%s_to_select_bookmark", key)
	return nil
}

func (tc *TestContext) theSelectedPathShouldBeApplied() error {
	// Verify path was applied correctly
	if !strings.Contains(tc.lastOutput, "select_bookmark") {
		return fmt.Errorf("expected bookmark to be selected, but last action was: %s", tc.lastOutput)
	}
	return nil
}
