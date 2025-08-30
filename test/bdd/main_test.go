package bdd

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/DamianReeves/sync-tools/test/bdd/steps"
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
	opts := godog.Options{
		Format:        "pretty",
		Paths:         []string{"../../features"},
		Randomize:     0,
		StopOnFailure: false,
		Tags:          "~@wip", // Exclude work-in-progress scenarios
	}

	suite := godog.TestSuite{
		Name:                 "sync-tools",
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}

	if suite.Run() != 0 {
		os.Exit(1)
	}
}