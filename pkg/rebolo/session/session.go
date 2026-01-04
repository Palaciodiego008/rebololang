package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionStore wraps gorilla/sessions Store
type SessionStore struct {
	store sessions.Store
	name  string
}

// NewCookieSessionStore creates a new cookie-based session store
func NewCookieSessionStore(name string, keyPairs ...[]byte) *SessionStore {
	store := sessions.NewCookieStore(keyPairs...)

	// Set secure defaults
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return &SessionStore{
		store: store,
		name:  name,
	}
}

// Session represents a user session
type Session struct {
	session *sessions.Session
	r       *http.Request
	w       http.ResponseWriter
}

// Get retrieves a session
func (ss *SessionStore) Get(r *http.Request, w http.ResponseWriter) (*Session, error) {
	session, err := ss.store.Get(r, ss.name)
	if err != nil {
		return nil, err
	}

	return &Session{
		session: session,
		r:       r,
		w:       w,
	}, nil
}

// Set stores a value in the session
func (s *Session) Set(key string, value interface{}) {
	s.session.Values[key] = value
}

// Get retrieves a value from the session
func (s *Session) Get(key string) interface{} {
	return s.session.Values[key]
}

// GetString retrieves a string value from the session
func (s *Session) GetString(key string) string {
	val, ok := s.session.Values[key].(string)
	if !ok {
		return ""
	}
	return val
}

// GetInt retrieves an int value from the session
func (s *Session) GetInt(key string) int {
	val, ok := s.session.Values[key].(int)
	if !ok {
		return 0
	}
	return val
}

// GetBool retrieves a bool value from the session
func (s *Session) GetBool(key string) bool {
	val, ok := s.session.Values[key].(bool)
	if !ok {
		return false
	}
	return val
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
	delete(s.session.Values, key)
}

// Clear removes all values from the session
func (s *Session) Clear() {
	for key := range s.session.Values {
		delete(s.session.Values, key)
	}
}

// Save persists the session
func (s *Session) Save() error {
	return s.session.Save(s.r, s.w)
}

// Destroy invalidates the session
func (s *Session) Destroy() error {
	s.session.Options.MaxAge = -1
	return s.session.Save(s.r, s.w)
}

// AddFlash adds a flash message to the session
func (s *Session) AddFlash(value interface{}, vars ...string) {
	s.session.AddFlash(value, vars...)
}

// Flashes retrieves and clears flash messages
func (s *Session) Flashes(vars ...string) []interface{} {
	return s.session.Flashes(vars...)
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.session.ID
}

// IsNew returns true if the session is new
func (s *Session) IsNew() bool {
	return s.session.IsNew
}
