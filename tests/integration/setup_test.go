package integration

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"weatherapi/tests/testapp"
)

var testServer *httptest.Server
var container *testapp.TestContainer

func TestMain(m *testing.M) {
	container = testapp.Initialize()

	// Log that testapp initialized successfully
	log.Printf("Testapp initialized with DB: %v", container.DB != nil)

	testServer = httptest.NewServer(container.Router)
	defer testServer.Close()

	code := m.Run()
	os.Exit(code)
}
