package main

import (
	"testing"
)

func TestVerifyApiToken(t *testing.T) {
	apiTokens = []string{"abcdefg"}
	err := VerifyApiToken("abcdefg")
	if err != nil {
		t.Errorf("VerifyApiToken failed with valid token")
	}
	err = VerifyApiToken("ABCDEFG")
	if err == nil {
		t.Errorf("VerifyApiToken success without valid token")
	}
}
