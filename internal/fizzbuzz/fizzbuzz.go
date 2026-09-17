// Package fizzbuzz implements the core "fizz-buzz" algorithm:
// for every integer from 1 to Limit, multiples of Int1 are replaced by Str1,
// multiples of Int2 are replaced by Str2, and multiples of both are replaced
// by the concatenation Str1+Str2. Every other number is rendered as-is.
package fizzbuzz

import (
	"errors"
	"strconv"
)

// MaxLimit caps how large a single request can be. Without a cap a client could request
// a huge number and force the server to allocate an unbounded amount of memory/CPU.
// 1000000 is large enough for any realistic use case.
const MaxLimit = 1000000

// MaxStrLen caps how long str1 and str2 can be.
// Without a cap a client could request a huge string
// causing unbounded memory growth across distinct requests.
// 100 is large enough for any realistic use case.
const MaxStrLen = 100

// Request holds the validated parameters for the fizzbuzz request.
type Request struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// validation error messages, to be called by the Validate() function
// distinguish "bad input" from unexpected failures.
var (
	ErrInt1MustBePositive = errors.New("int1 must be a positive integer")
	ErrInt2MustBePositive = errors.New("int2 must be a positive integer")
	ErrLimitRange         = errors.New("limit must be between 1 and " + strconv.Itoa(MaxLimit))
	ErrStr1Empty          = errors.New("str1 must not be empty")
	ErrStr2Empty          = errors.New("str2 must not be empty")
	ErrStr1TooLong        = errors.New("str1 must be no longer than " + strconv.Itoa(MaxStrLen) + " characters")
	ErrStr2TooLong        = errors.New("str2 must be no longer than " + strconv.Itoa(MaxStrLen) + " characters")
)

// Validate checks if the request parameters are valid.
func (r Request) Validate() error {
	switch {
	case r.Int1 <= 0:
		return ErrInt1MustBePositive
	case r.Int2 <= 0:
		return ErrInt2MustBePositive
	case r.Limit < 1 || r.Limit > MaxLimit:
		return ErrLimitRange
	case r.Str1 == "":
		return ErrStr1Empty
	case r.Str2 == "":
		return ErrStr2Empty
	case len(r.Str1) > MaxStrLen:
		return ErrStr1TooLong
	case len(r.Str2) > MaxStrLen:
		return ErrStr2TooLong
	}
	return nil
}

// Generate returns a slice of strings representing the fizzbuzz sequence for the request.
// Callers are expected to call Validate first and only pass valid requests to Generate.
func Generate(r Request) []string {
	result := make([]string, r.Limit)

	for i := 1; i <= r.Limit; i++ {
		switch {
		case i%r.Int1 == 0 && i%r.Int2 == 0:
			result[i-1] = r.Str1 + r.Str2
		case i%r.Int1 == 0:
			result[i-1] = r.Str1
		case i%r.Int2 == 0:
			result[i-1] = r.Str2
		default:
			result[i-1] = strconv.Itoa(i)
		}
	}
	return result
}
