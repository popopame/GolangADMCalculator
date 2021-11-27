package main

import "testing"

func TestComputeMomentum(t *testing.T) {

	testHistory := []float64{469.185, 468.19, 459.25, 429.14, 451.56, 438.51, 428.06, 420.04, 417.3, 396.33, 380.36, 370.07}
	testComputedMomentum := 5.6975

	computedMomentum, err := ComputeMomentum(testHistory)

	if err != nil {
		t.Errorf("Function TestComputedMomentum returned an error: %v", err)
	}

	if computedMomentum != testComputedMomentum {
		t.Errorf("Error in computing the Momentum, wanted %v, got %v", testComputedMomentum, computedMomentum)
	}

}
