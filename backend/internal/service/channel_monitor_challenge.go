package service

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
)

// Arithmetic few-shot prompt template for monitor challenges.
const monitorChallengePromptTemplate = `Calculate and respond with ONLY the number, nothing else.

Q: 3 + 5 = ?
A: 8

Q: 12 - 7 = ?
A: 5

Q: %d %s %d = ?
A:`

var monitorChallengeNumberRegex = regexp.MustCompile(`-?\d+`)

type monitorChallenge struct {
	Prompt   string
	Expected string
}

// generateChallenge creates a random arithmetic challenge with a few-shot
// prompt and the expected answer string.
func generateChallenge() monitorChallenge {
	x := randIntInRange(monitorChallengeMin, monitorChallengeMax)
	y := randIntInRange(monitorChallengeMin, monitorChallengeMax)

	useAddition := rand.IntN(2) == 0 //nolint:gosec
	if useAddition {
		return monitorChallenge{
			Prompt:   fmt.Sprintf(monitorChallengePromptTemplate, x, "+", y),
			Expected: strconv.Itoa(x + y),
		}
	}

	larger, smaller := x, y
	if smaller > larger {
		larger, smaller = smaller, larger
	}
	return monitorChallenge{
		Prompt:   fmt.Sprintf(monitorChallengePromptTemplate, larger, "-", smaller),
		Expected: strconv.Itoa(larger - smaller),
	}
}

func randIntInRange(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + rand.IntN(hi-lo+1) //nolint:gosec
}

// validateChallenge checks whether the expected integer appears anywhere in
// the response text.
func validateChallenge(response, expected string) bool {
	if response == "" || expected == "" {
		return false
	}
	for _, num := range monitorChallengeNumberRegex.FindAllString(response, -1) {
		if num == expected {
			return true
		}
	}
	return false
}
