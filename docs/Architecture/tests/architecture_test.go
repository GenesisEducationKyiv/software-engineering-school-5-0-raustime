package architecture_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

func TestSubscriptionService_ShouldNotDependOnPresentation(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/services/subscription_service").
		ShouldNotDependOn("weatherapi/internal/server")
}

func TestWeatherService_ShouldNotDependOnPresentation(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/services/weather_service").
		ShouldNotDependOn("weatherapi/internal/server")
}

func TestMailerService_ShouldNotDependOnPresentation(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/services/mailer_service").
		ShouldNotDependOn("weatherapi/internal/server")
}

func TestHandlers_ShouldNotDependDirectlyOnAdapters(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/server/handlers").
		ShouldNotDependDirectlyOn("weatherapi/internal/adapters")
}

func TestMiddleware_ShouldNotDependOnBusinessLogic(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/server/middleware").
		ShouldNotDependOn("weatherapi/internal/services")
}

func TestDBRepositories_ShouldNotDependOnServer(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/db/repositories").
		ShouldNotDependOn("weatherapi/internal/server")
}

func TestCache_ShouldNotDependOnBusinessLogic(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/cache").
		ShouldNotDependOn("weatherapi/internal/services/weather_service")
}

func TestContracts_ShouldRemainIsolated(t *testing.T) {
	archtest.Package(t, "weatherapi/internal/contracts").
		ShouldNotDependOn(
			"weatherapi/internal/services",
			"weatherapi/internal/server",
			"weatherapi/internal/adapters",
			"weatherapi/internal/cache",
			"weatherapi/internal/db",
		)
}
