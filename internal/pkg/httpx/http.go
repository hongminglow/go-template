package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxBodySize = 1 << 20 // 1MB

// Problem represents an RFC 7807 error response.
type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Errors   map[string]any `json:"errors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"title":"Internal Server Error","status":500}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(append(payload, '\n'))
}

// WriteProblem writes an RFC 7807 problem response.
func WriteProblem(w http.ResponseWriter, statusCode int, title string, detail string) {
	WriteJSON(w, statusCode, Problem{
		Type:   fmt.Sprintf("https://httpstatuses.com/%d", statusCode),
		Title:  title,
		Status: statusCode,
		Detail: detail,
	})
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteProblem(w, statusCode, http.StatusText(statusCode), message)
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
