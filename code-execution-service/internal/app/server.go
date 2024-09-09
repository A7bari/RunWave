package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/A7bari/RunWave/internal"
)

type Server struct {
	podManager *internal.PodManager
}

func NewServer(podManager *internal.PodManager) *Server {
	return &Server{
		podManager: podManager,
	}
}

func (s *Server) Start() {

	// code exection handler
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		var req internal.CodeExecutionReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		output, err := s.podManager.ExecuteCode(req.Code, req.Language)

		resp := &internal.CodeExecutionResp{
			Output:     output,
			StatusCode: http.StatusOK,
		}

		if err != nil {
			handleErrors(err, resp)
			log.Printf("Error: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to send response: %v", err)
		}

	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		data := s.podManager.HealthzCheck()

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ErrorHandler handles different types of errors and sends appropriate HTTP responses
// some error are related to code execution
func handleErrors(err error, resp *internal.CodeExecutionResp) {
	var ierr *internal.Error
	resp.Output = ""

	if errors.As(err, &ierr) {
		switch ierr.Code() {
		// code execution errors
		case internal.ErrorCodeExecutionErr:
			resp.Error = err.Error()
			return
		case internal.ErrorCodeTimeout:
			resp.Error = "Execution timeout!"

		// internal errors
		case internal.ErrorCodePodNotFound:
			resp.StatusCode = http.StatusServiceUnavailable
			resp.Error = "The service is currently unavailable, try again later!"
		case internal.ErrorCodeUnsupportLanguage:
			resp.StatusCode = http.StatusBadRequest
			resp.Error = "Unsupported language!"
		default:
			resp.StatusCode = http.StatusInternalServerError
			resp.Error = "Internal server error"
		}
	} else {
		resp.StatusCode = http.StatusInternalServerError
		resp.Error = "Internal server error"
	}
}
