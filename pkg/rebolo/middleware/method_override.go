package middleware

import (
	"net/http"
)

// MethodOverride middleware allows HTML forms to use PUT, PATCH, and DELETE methods
// by checking the _method form value. This is needed because HTML forms only support GET and POST.
// IMPORTANT: This middleware must run BEFORE the router processes the request.
func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process POST requests
		if r.Method == "POST" {
			// Try to parse form to get _method field
			// ParseForm() will parse both URL query and POST body
			_ = r.ParseForm()

			// Get _method from form values (works with both Form and PostForm)
			method := r.FormValue("_method")
			if method == "" {
				// Also try PostFormValue as fallback
				method = r.PostFormValue("_method")
			}

			if method != "" {
				switch method {
				case "PUT", "PATCH", "DELETE":
					r.Method = method
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
