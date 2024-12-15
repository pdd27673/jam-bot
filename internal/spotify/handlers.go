// internal/spotify/handlers.go
package spotify

import (
	"fmt"
	"net/http"
)

func (s *Service) StartAuthServer(port int) error {
	http.HandleFunc("/callback", s.callbackHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *Service) callbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if err := s.HandleCallback(r.Context(), state, code); err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Close window after successful auth
	fmt.Fprintf(w, "<script>window.close()</script>")
}
