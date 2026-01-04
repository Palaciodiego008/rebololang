package session

import (
	"net/http"
)

// GetSession is a convenience function to get session from request context
// Usage in controllers: session, _ := rebolo.GetSession(r, w)
func GetSession(r *http.Request, w http.ResponseWriter) (*Session, error) {
	// Get the application instance from context (if available)
	// For now, we'll create a default session store
	// This can be improved by storing app in context
	store := NewCookieSessionStore("rebolo_session", []byte("rebolo-secret-key-change-in-production"))
	return store.Get(r, w)
}

// GetFlash is a convenience function to get flash messages
// Usage in controllers: flash := rebolo.GetFlash(r, w)
func GetFlash(r *http.Request, w http.ResponseWriter) *Flash {
	session, err := GetSession(r, w)
	if err != nil {
		return &Flash{session: &Session{}}
	}
	return NewFlash(session)
}
