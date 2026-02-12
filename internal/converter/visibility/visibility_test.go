package visibility

import (
	"testing"

	api_visibility "google.golang.org/genproto/googleapis/api/visibility"
)

func TestShouldBeFiltered(t *testing.T) {
	tests := []struct {
		name       string
		rule       *api_visibility.VisibilityRule
		selectors  map[string]bool
		wantFilter bool
	}{
		{
			name:       "nil rule is never filtered",
			rule:       nil,
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "rule with empty selectors is always filtered",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL"},
			selectors:  map[string]bool{},
			wantFilter: true,
		},
		{
			name:       "single restriction matching selector",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "single restriction not matching selector",
			rule:       &api_visibility.VisibilityRule{Restriction: "EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: true,
		},
		{
			name:       "comma-separated restriction with first label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with second label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"EXTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with no label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"PREVIEW": true},
			wantFilter: true,
		},
		{
			name:       "comma-separated restriction with spaces",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL, EXTERNAL"},
			selectors:  map[string]bool{"EXTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with multiple selectors",
			rule:       &api_visibility.VisibilityRule{Restriction: "PREVIEW,EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true, "PREVIEW": true},
			wantFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldBeFiltered(tt.rule, tt.selectors)
			if got != tt.wantFilter {
				t.Errorf("ShouldBeFiltered() = %v, want %v", got, tt.wantFilter)
			}
		})
	}
}
