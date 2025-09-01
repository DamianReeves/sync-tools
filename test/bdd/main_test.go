package bdd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DamianReeves/sync-tools/test/bdd/steps"
	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features"},
			TestingT: t,
			Tags:     "~@wip", // Exclude work-in-progress scenarios
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	testContext := steps.NewTestContext()
	testContext.RegisterSteps(ctx)
}

func TestMain(m *testing.M) {
	// Build the sync-tools binary before running tests
	if err := buildSyncToolsBinary(); err != nil {
		os.Exit(1)
	}

	opts := godog.Options{
		Format:        "pretty",
		Paths:         []string{"../../features"},
		Randomize:     0,
		StopOnFailure: false,
		Tags:          "~@wip", // Exclude work-in-progress scenarios
	}

	suite := godog.TestSuite{
		Name:                "sync-tools",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}

	if suite.Run() != 0 {
		os.Exit(1)
	}
}

// buildSyncToolsBinary builds the sync-tools binary required for BDD tests
func buildSyncToolsBinary() error {
	// Get the project root directory (two levels up from test/bdd)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		return err
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "sync-tools", "./cmd/sync-tools")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
