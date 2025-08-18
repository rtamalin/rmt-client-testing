package choice

import (
	"math/rand"
	"time"
)

type Choice struct {
	Weight int
	Value  any
}

func randomValue(numValues int) int {
	rs := rand.NewSource(time.Now().UnixNano())

	r := rand.New(rs)

	return r.Intn(numValues)
}

func Choose(choices []Choice) any {
	var totWeight int
	var weights []int

	// determine total of all weights, and a weighting table
	for _, choice := range choices {
		totWeight += choice.Weight
		weights = append(weights, totWeight)
	}

	randChoice := randomValue(totWeight)

	var chosen int
	for i, weight := range weights {
		if randChoice < weight {
			chosen = i
			break
		}
	}

	// return chosen value
	return choices[chosen].Value
}
