package fizzbuzz

import (
	"reflect"
	"strings"
	"testing"
)

func TestGenerate_Classic(t *testing.T) {
	//Multiplies of 3 -> "fizz"
	//Multiplies of 5 -> "buzz"
	//Multiplies of 3 and 5 -> "fizzbuzz"
	//Others -> number
	req := Request{
		Int1:  3,
		Int2:  5,
		Limit: 16,
		Str1:  "fizz",
		Str2:  "buzz",
	}

	got := Generate(req)
	want := []string{
		"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz",
		"11", "fizz", "13", "14", "fizzbuzz", "16",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Generate()= got: %v, want: %v", got, want)
	}
}

func TestGenerate_CustomParams(t *testing.T) {
	req := Request{Int1: 2, Int2: 7, Limit: 20, Str1: "Hel", Str2: "my"}
	got := Generate(req)
	want := []string{
		"1", "Hel", "3", "Hel", "5", "Hel", "my", "Hel",
		"9", "Hel", "11", "Hel", "13", "Helmy", "15", "Hel",
		"17", "Hel", "19", "Hel",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Generate()= got: %v, want: %v", got, want)
	}
}

// Assuming that fizzbuzz has priority over fizz and buzz when int1 and int2 are equal.
func TestGenerate_BothDivisorsEqual(t *testing.T) {
	req := Request{Int1: 1, Int2: 1, Limit: 5, Str1: "fizz", Str2: "buzz"}
	got := Generate(req)
	want := []string{"fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Generate() = %v, want %v", got, want)
	}
}

// TestGenerate_LimitOne asserts the boundary where limit=1 and the single
// element (1) is not a multiple of either divisor.
func TestGenerate_LimitOne(t *testing.T) {
	req := Request{Int1: 3, Int2: 5, Limit: 1, Str1: "fizz", Str2: "buzz"}
	got := Generate(req)
	want := []string{"1"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("Generate() = %v, want %v", got, want)
	}
}

// TestGenerate_Int1IsOne confirms that when int1=1 every number becomes str1
// (or str1+str2 for multiples of int2).
func TestGenerate_Int1IsOne(t *testing.T) {
	req := Request{Int1: 1, Int2: 4, Limit: 8, Str1: "fizz", Str2: "buzz"}
	got := Generate(req)
	// Every number is a multiple of 1 -> str1; multiples of 4 also get str2.
	want := []string{"fizz", "fizz", "fizz", "fizzbuzz", "fizz", "fizz", "fizz", "fizzbuzz"}
	if len(got) != len(want) {
		t.Fatalf("Generate() len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Generate()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// Table-driven tests for Validate function.
func TestValidate(t *testing.T) {
	longStr := strings.Repeat("a", MaxStrLen+1)

	tests := []struct {
		testName string
		req      Request
		wantErr  error
	}{
		{"valid", Request{3, 5, 15, "fizz", "buzz"}, nil},
		{"int1 zero", Request{0, 5, 15, "fizz", "buzz"}, ErrInt1MustBePositive},
		{"int1 negative", Request{-1, 5, 15, "fizz", "buzz"}, ErrInt1MustBePositive},
		{"int2 zero", Request{3, 0, 15, "fizz", "buzz"}, ErrInt2MustBePositive},
		{"int2 negative", Request{3, -1, 15, "fizz", "buzz"}, ErrInt2MustBePositive},
		{"limit zero", Request{3, 5, 0, "fizz", "buzz"}, ErrLimitRange},
		{"limit negative", Request{3, 5, -5, "fizz", "buzz"}, ErrLimitRange},
		{"limit too large", Request{3, 5, MaxLimit + 1, "fizz", "buzz"}, ErrLimitRange},
		{"str1 empty", Request{3, 5, 15, "", "buzz"}, ErrStr1Empty},
		{"str1 too long", Request{3, 5, 15, longStr, "buzz"}, ErrStr1TooLong},
		{"str2 empty", Request{3, 5, 15, "fizz", ""}, ErrStr2Empty},
		{"str2 too long", Request{3, 5, 15, "fizz", longStr}, ErrStr2TooLong},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := tt.req.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
