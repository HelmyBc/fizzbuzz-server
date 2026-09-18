package server

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/HelmyBc/fizzbuzz-server/internal/stats"
)

// statisticsResponse is the JSON representation of the most-used FizzBuzz parameters.
type statisticsResponse struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
	Hits  int64  `json:"hits"`
}

// StatisticsHandler returns statistics for the most-used FizzBuzz parameters.
type StatisticsHandler struct {
	Stats *stats.Store
}

func (h StatisticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	entry, ok := h.Stats.Most()
	if !ok {
		writeError(w, http.StatusNotFound, "no requests have been recorded yet")
		return
	}

	params, err := url.ParseQuery(entry.Key)
	if err != nil {
		slog.Error("failed to parse stats key", "key", entry.Key, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error : failed to parse statistics")
		return
	}

	parseInt := func(name string) int {
		n, _ := strconv.Atoi(params.Get(name))
		return n
	}

	writeJSON(w, http.StatusOK, statisticsResponse{
		Int1:  parseInt("int1"),
		Int2:  parseInt("int2"),
		Limit: parseInt("limit"),
		Str1:  params.Get("str1"),
		Str2:  params.Get("str2"),
		Hits:  entry.Count,
	})
}
