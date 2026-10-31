package domain

import (
	"errors"
	"testing"
)

func TestOpeningValidate(t *testing.T) {
	tests := []struct {
		name    string
		opening Opening
		valid   bool
	}{
		{name: "valid", opening: Opening{Role: "Engineer", Company: "Acme", Link: "https://acme.test/jobs/1"}, valid: true},
		{name: "missing role", opening: Opening{Company: "Acme"}},
		{name: "negative salary", opening: Opening{Role: "Engineer", Company: "Acme", Salary: -1}},
		{name: "relative link", opening: Opening{Role: "Engineer", Company: "Acme", Link: "/jobs/1"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.opening.Validate()
			if test.valid && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !test.valid && !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}
