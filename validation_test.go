package main

import "testing"

func TestValidateStatus(t *testing.T) {
	validStatuses := []string{"pendiente", "jugando", "completado", "abandonado"}
	for _, status := range validStatuses {
		if !ValidateStatus(status) {
			t.Fatalf("ValidateStatus(%q) = false, want true", status)
		}
	}

	invalidStatuses := []string{"pendiente ", "playing", "done", "", "invalido"}
	for _, status := range invalidStatuses {
		if ValidateStatus(status) {
			t.Fatalf("ValidateStatus(%q) = true, want false", status)
		}
	}
}

func TestValidatePersonalScore(t *testing.T) {
	validScores := []int{1, 5, 10}
	for _, score := range validScores {
		if !ValidatePersonalScore(score) {
			t.Fatalf("ValidatePersonalScore(%d) = false, want true", score)
		}
	}

	invalidScores := []int{0, -1, 11, 100}
	for _, score := range invalidScores {
		if ValidatePersonalScore(score) {
			t.Fatalf("ValidatePersonalScore(%d) = true, want false", score)
		}
	}
}
