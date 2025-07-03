package architecture_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

// Presentation layer не повинен напряму залежати від інфраструктури
func TestPresentationLayerShouldNotDependOnInfrastructure(t *testing.T) {
	archtest.Package("weatherapi/internal/server").
		ShouldNotDependOn(
			"weatherapi/internal/adapters",
			"weatherapi/internal/cache",
			"weatherapi/internal/services/mailer_service",
			"weatherapi/internal/logging",
		).
		FailIfInvalid(t)
}

// Application layer не повинен залежати від HTTP або middleware
func TestApplicationLayerShouldNotDependOnPresentation(t *testing.T) {
	archtest.Package("weatherapi/internal/services").
		ShouldNotDependOn("weatherapi/internal/server").
		FailIfInvalid(t)
}

// Domain layer (contracts) повинен залежати лише від стандартної бібліотеки
func TestDomainShouldHaveOnlyStandardDependencies(t *testing.T) {
	archtest.Package("weatherapi/internal/contracts").
		ShouldOnlyDependOn(
			"context", "errors", "time", // дозволені з stdlib
		).
		FailIfInvalid(t)
}
