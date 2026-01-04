package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// AppRouter interface for testing (to avoid circular dependencies)
type AppRouter interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// TestApp provides testing utilities
type TestApp struct {
	router AppRouter
	Server *httptest.Server
}

// NewTestApp creates a new test application with the provided router
func NewTestApp(router AppRouter) *TestApp {
	return &TestApp{
		router: router,
	}
}

// StartServer starts a test HTTP server
func (ta *TestApp) StartServer() {
	if ta.Server == nil {
		ta.Server = httptest.NewServer(ta.router)
	}
}

// StopServer stops the test HTTP server
func (ta *TestApp) StopServer() {
	if ta.Server != nil {
		ta.Server.Close()
		ta.Server = nil
	}
}

// Router returns the app router
func (ta *TestApp) Router() AppRouter {
	return ta.router
}

// TestRequest represents a test HTTP request
type TestRequest struct {
	method  string
	path    string
	body    io.Reader
	headers map[string]string
	cookies []*http.Cookie
	app     *TestApp
}

// NewTestRequest creates a new test request
func (ta *TestApp) NewRequest(method, path string) *TestRequest {
	return &TestRequest{
		method:  method,
		path:    path,
		headers: make(map[string]string),
		cookies: make([]*http.Cookie, 0),
		app:     ta,
	}
}

// GET creates a GET request
func (ta *TestApp) GET(path string) *TestRequest {
	return ta.NewRequest("GET", path)
}

// POST creates a POST request
func (ta *TestApp) POST(path string) *TestRequest {
	return ta.NewRequest("POST", path)
}

// PUT creates a PUT request
func (ta *TestApp) PUT(path string) *TestRequest {
	return ta.NewRequest("PUT", path)
}

// DELETE creates a DELETE request
func (ta *TestApp) DELETE(path string) *TestRequest {
	return ta.NewRequest("DELETE", path)
}

// PATCH creates a PATCH request
func (ta *TestApp) PATCH(path string) *TestRequest {
	return ta.NewRequest("PATCH", path)
}

// WithHeader adds a header to the request
func (tr *TestRequest) WithHeader(key, value string) *TestRequest {
	tr.headers[key] = value
	return tr
}

// WithCookie adds a cookie to the request
func (tr *TestRequest) WithCookie(cookie *http.Cookie) *TestRequest {
	tr.cookies = append(tr.cookies, cookie)
	return tr
}

// WithJSON sets the request body as JSON
func (tr *TestRequest) WithJSON(data interface{}) *TestRequest {
	jsonData, _ := json.Marshal(data)
	tr.body = bytes.NewBuffer(jsonData)
	tr.headers["Content-Type"] = "application/json"
	return tr
}

// WithForm sets the request body as form data
func (tr *TestRequest) WithForm(data map[string]string) *TestRequest {
	form := url.Values{}
	for key, value := range data {
		form.Set(key, value)
	}
	tr.body = strings.NewReader(form.Encode())
	tr.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return tr
}

// WithBody sets the request body
func (tr *TestRequest) WithBody(body io.Reader) *TestRequest {
	tr.body = body
	return tr
}

// Do executes the request and returns the response
func (tr *TestRequest) Do() *TestResponse {
	req := httptest.NewRequest(tr.method, tr.path, tr.body)

	// Add headers
	for key, value := range tr.headers {
		req.Header.Set(key, value)
	}

	// Add cookies
	for _, cookie := range tr.cookies {
		req.AddCookie(cookie)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	tr.app.router.ServeHTTP(w, req)

	return &TestResponse{
		ResponseRecorder: w,
		request:          req,
	}
}

// TestResponse wraps httptest.ResponseRecorder with helper methods
type TestResponse struct {
	*httptest.ResponseRecorder
	request *http.Request
}

// Status returns the HTTP status code
func (tr *TestResponse) Status() int {
	return tr.Code
}

// Body returns the response body as string
func (tr *TestResponse) Body() string {
	return tr.ResponseRecorder.Body.String()
}

// BodyBytes returns the response body as bytes
func (tr *TestResponse) BodyBytes() []byte {
	return tr.ResponseRecorder.Body.Bytes()
}

// JSON decodes the response body as JSON
func (tr *TestResponse) JSON(v interface{}) error {
	return json.Unmarshal(tr.BodyBytes(), v)
}

// Header returns a response header
func (tr *TestResponse) Header(key string) string {
	return tr.ResponseRecorder.Header().Get(key)
}

// Cookie returns a response cookie
func (tr *TestResponse) Cookie(name string) *http.Cookie {
	for _, cookie := range tr.ResponseRecorder.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// IsOK returns true if status is 200
func (tr *TestResponse) IsOK() bool {
	return tr.Code == http.StatusOK
}

// IsCreated returns true if status is 201
func (tr *TestResponse) IsCreated() bool {
	return tr.Code == http.StatusCreated
}

// IsRedirect returns true if status is 3xx
func (tr *TestResponse) IsRedirect() bool {
	return tr.Code >= 300 && tr.Code < 400
}

// IsClientError returns true if status is 4xx
func (tr *TestResponse) IsClientError() bool {
	return tr.Code >= 400 && tr.Code < 500
}

// IsServerError returns true if status is 5xx
func (tr *TestResponse) IsServerError() bool {
	return tr.Code >= 500 && tr.Code < 600
}

// Contains checks if the response body contains a substring
func (tr *TestResponse) Contains(substr string) bool {
	return strings.Contains(tr.Body(), substr)
}

// ContainsAll checks if the response body contains all substrings
func (tr *TestResponse) ContainsAll(substrs ...string) bool {
	body := tr.Body()
	for _, substr := range substrs {
		if !strings.Contains(body, substr) {
			return false
		}
	}
	return true
}
