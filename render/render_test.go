package render

import (
	"strings"
	"testing"

	"copycat/move"
)

func TestMapInput(t *testing.T) {
	cases := []struct {
		name              string
		left, right, jump bool
		want              move.Input
	}{
		{name: "left alone", left: true, want: move.Input{Steer: -1}},
		{name: "right alone", right: true, want: move.Input{Steer: 1}},
		{name: "both cancel", left: true, right: true, want: move.Input{Steer: 0}},
		{name: "jump alone", jump: true, want: move.Input{Steer: 0, Jump: true}},
		{name: "left and jump", left: true, jump: true, want: move.Input{Steer: -1, Jump: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MapInput(tc.left, tc.right, tc.jump)
			if got != tc.want {
				t.Errorf("MapInput(left=%v, right=%v, jump=%v) = %+v, want %+v",
					tc.left, tc.right, tc.jump, got, tc.want)
			}
		})
	}
}

func TestReelsText(t *testing.T) {
	cases := []struct {
		found, total int
		want         string
	}{
		{0, 20, "0/20"},
		{12, 20, "12/20"},
		{20, 20, "20/20"},
	}
	for _, tc := range cases {
		if got := ReelsText(tc.found, tc.total); got != tc.want {
			t.Errorf("ReelsText(%d, %d) = %q, want %q", tc.found, tc.total, got, tc.want)
		}
	}
}

func TestTakeText(t *testing.T) {
	if got := TakeText(); got != "Take 9" {
		t.Errorf("TakeText() = %q, want %q", got, "Take 9")
	}
}

func TestRecapText(t *testing.T) {
	if got := RecapText(12, 20); got != "Reels: 12 of 20" {
		t.Errorf("RecapText(12, 20) = %q, want %q", got, "Reels: 12 of 20")
	}
}

func TestTextsPassToneMap(t *testing.T) {
	// Representative terms for every banned category: blood, gore,
	// dismemberment, sexual content, drugs, profanity, slurs and jokes against
	// real groups or religions.
	banned := []string{
		"blood", "gore", "dismember", "corpse", "decapitat",
		"sex", "nudity", "porn", "rape",
		"drug", "cocaine", "heroin", "meth", "opium",
		"fuck", "shit", "bitch", "bastard", "damn", "asshole", "cunt",
		"nigger", "faggot", "retard", "cripple",
		"allah", "jesus", "christ", "buddha", "muslim", "jew", "hindu",
	}
	texts := []struct {
		name string
		text string
	}{
		{name: "game name", text: gameName},
		{name: "start prompt", text: startPrompt},
		{name: "retry prompt", text: retryPrompt},
		{name: "leave prompt", text: leavePrompt},
	}
	for _, tc := range texts {
		if strings.TrimSpace(tc.text) == "" {
			t.Errorf("%s is empty", tc.name)
		}
		lower := strings.ToLower(tc.text)
		for _, term := range banned {
			if strings.Contains(lower, term) {
				t.Errorf("%s %q contains banned term %q", tc.name, tc.text, term)
			}
		}
	}
}
