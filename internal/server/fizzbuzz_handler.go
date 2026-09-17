package server

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/HelmyBc/fizzbuzz-server/internal/fizzbuzz"
	"github.com/HelmyBc/fizzbuzz-server/internal/stats"
)

// fizzBuzzResponse is the JSON body returned by a successful /fizzbuzz call.
type fizzBuzzResponse struct {
	Result []string `json:"result"`
}

// FizzBuzzHandler serves GET /fizzbuzz. Every valid request is also recorded
// in the stats store so the /statistics endpoint can report on it.
type FizzBuzzHandler struct {
	Stats *stats.Store
}

func (h *FizzBuzzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := parseRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.Stats.Increment(statsKey(req))

	result := fizzbuzz.Generate(req)

	writeJSON(w, http.StatusOK, fizzBuzzResponse{Result: result})
}

// parseRequest extracts and parses the fizz-buzz parameters from the query string.
// It returns an error when a required integer parameter is missing or invalid.
func parseRequest(r *http.Request) (fizzbuzz.Request, error) {
	q := r.URL.Query()

	int1, err := parseIntParam(q, "int1")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	int2, err := parseIntParam(q, "int2")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	limit, err := parseIntParam(q, "limit")
	if err != nil {
		return fizzbuzz.Request{}, err
	}

	str1 := q.Get("str1")
	str2 := q.Get("str2")

	return fizzbuzz.Request{
		Int1:  int1,
		Int2:  int2,
		Limit: limit,
		Str1:  str1,
		Str2:  str2,
	}, nil
}

// parseIntParam retrieves a required integer query parameter.
// It returns a descriptive error when the parameter is missing or invalid.
func parseIntParam(q map[string][]string, name string) (int, error) {
	values, ok := q[name]
	if !ok || len(values) == 0 || values[0] == "" {
		return 0, errors.New(name + " is required")
	}
	n, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, errors.New(name + " must be a valid integer")
	}
	return n, nil
}

// statsKey returns a canonical representation of a fizz-buzz request.
// The encoded query format ensures that equivalent requests produce the same key.
func statsKey(r fizzbuzz.Request) string {
	v := url.Values{}
	v.Set("int1", strconv.Itoa(r.Int1))
	v.Set("int2", strconv.Itoa(r.Int2))
	v.Set("limit", strconv.Itoa(r.Limit))
	v.Set("str1", r.Str1)
	v.Set("str2", r.Str2)
	return v.Encode()
}
