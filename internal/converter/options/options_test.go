package options

import (
	"testing"
)

func TestOnlyGoogleapiHTTPOptionParsing(t *testing.T) {
	// Test that only-googleapi-http option is parsed correctly
	opts, err := FromString("only-googleapi-http")
	if err != nil {
		t.Fatalf("Failed to parse only-googleapi-http option: %v", err)
	}
	
	if !opts.OnlyGoogleapiHTTP {
		t.Error("Expected OnlyGoogleapiHTTP to be true when only-googleapi-http option is provided")
	}
	
	// Test that default value is false
	opts2, err := FromString("")
	if err != nil {
		t.Fatalf("Failed to parse empty options: %v", err)
	}
	
	if opts2.OnlyGoogleapiHTTP {
		t.Error("Expected OnlyGoogleapiHTTP to be false by default")
	}
	
	// Test combination with other options
	opts3, err := FromString("only-googleapi-http,debug")
	if err != nil {
		t.Fatalf("Failed to parse combined options: %v", err)
	}
	
	if !opts3.OnlyGoogleapiHTTP {
		t.Error("Expected OnlyGoogleapiHTTP to be true in combined options")
	}
	
	if !opts3.Debug {
		t.Error("Expected Debug to be true in combined options")
	}
}