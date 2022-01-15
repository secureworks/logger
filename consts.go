package log

// FIXME(PH): distribute to entry definition file and middleware.

const (
	// ReqDuration is a key for Logger data.
	ReqDuration = "request_duration"

	// PanicStack is a key for Logger data.
	PanicStack = "panic_stack"

	// PanicValue is a key for Logger data.
	PanicValue = "panic_value"

	// CallerField is a key for Logger data.
	CallerField = "caller"

	// StackField is a key for Logger data.
	StackField = "stack"

	// XRequestID is a common field and header from API calls.
	XRequestID = "X-Request-Id"

	// XTraceID is a common field and header from API calls.
	XTraceID = "X-Trace-Id"

	// XSpanID is a common field and header from API calls.
	XSpanID = "X-Span-Id"

	// XTenantCtx is a common field and header from API calls "within" a
	// tenant.
	XTenantCtx = "X-Tenant-Context"

	// XEnvironment is a header for API calls to services which span
	// environments (tenants).
	XEnvironment = "X-Environment"
)
