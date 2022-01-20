package logger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/middleware"
	_ "github.com/secureworks/logger/zerolog"
)

// You can easily integrate the Logger itself into HTTP server
// middleware too. See the middleware package for documentation
// examples.
func Example_usingMiddleware() {
	config := log.DefaultConfig(nil)
	config.Output = os.Stdout
	logger, _ := log.Open("zerolog", config)

	// Pick attributes to log. You can also skip defaults.
	attrs := &middleware.HTTPRequestLogAttributes{
		Headers:        []string{"X-Request-Id"},
		SkipDuration:   true,
		SkipRemoteAddr: true,
	}

	// Inject the logger and attributes into the middleware.
	mwareFn := middleware.NewHTTPRequestMiddleware(logger, log.INFO, attrs)
	handler := mwareFn(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	fmt.Println()

	// Send a request.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test/path", nil)
	req.Header.Add("X-Request-Id", "12345")
	srv.Client().Do(req)

	// Output:
	// {"http_method":"GET","http_path":"/test/path","x-request-id":"12345","level":"info"}
}
