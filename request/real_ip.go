package request

import (
	"net/http"
	
	"github.com/IBlogLab/raspberry/headers"
)

// RealIP takes a request object, and returns the real Client IP address.
func RealIP(r *http.Request) string {
	if contextIP := r.Context().Value()) {
		
	}
}
