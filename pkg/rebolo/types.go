package rebolo

// Re-export types from sub-packages for convenience
import (
	"net/http"

	"github.com/Palaciodiego008/rebololang/pkg/rebolo/context"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/errors"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/middleware"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/session"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/testing"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/validation"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/watcher"
)

// Type aliases for convenience
type (
	Context          = context.Context
	ContextHandler   = context.ContextHandler
	Session          = session.Session
	SessionStore     = session.SessionStore
	Flash            = session.Flash
	FlashMessage     = session.FlashMessage
	ErrorHandler     = errors.ErrorHandler
	ErrorHandlers    = errors.ErrorHandlers
	MiddlewareFunc   = middleware.MiddlewareFunc
	MiddlewareConfig = middleware.MiddlewareConfig
	MiddlewareStack  = middleware.MiddlewareStack
	FileWatcher      = watcher.FileWatcher
	TestApp          = testing.TestApp
	ValidationError  = validation.ValidationError
	ValidationErrors = validation.ValidationErrors
)

// Function aliases for convenience
var (
	NewContext            = context.NewContext
	NewCookieSessionStore = session.NewCookieSessionStore
	NewFlash              = session.NewFlash
	GetSession            = session.GetSession
	GetFlash              = session.GetFlash
	NewErrorHandlers      = errors.NewErrorHandlers
	NewMiddlewareStack    = middleware.NewMiddlewareStack
	CORSMiddleware        = middleware.CORSMiddleware
	ValidateStruct        = validation.ValidateStruct
	ValidationErrorsToMap = validation.ValidationErrorsToMap
	Bind                  = validation.Bind
	BindAndValidate       = validation.BindAndValidate
)

// NewTestApp creates a new test app wrapping an application
func NewTestApp(app *Application) *TestApp {
	return testing.NewTestApp(app.router)
}

// ContextMiddleware wraps a ContextHandler to work with standard http.Handler
func (a *Application) ContextMiddleware(handler ContextHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, a)

		if err := handler(ctx); err != nil {
			// Use custom error handler
			a.InternalErrorHandler(w, r, err)
		}
	}
}
