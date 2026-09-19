package main

import "testing"

func TestRunSubscriptionUpdateRecoversPanic(t *testing.T) {
	if !runSubscriptionUpdate(3, func() { panic("malformed plugin") }) {
		t.Fatal("panic was not recovered")
	}
}
