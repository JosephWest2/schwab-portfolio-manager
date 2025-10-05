package util

import (
	"math"
)

func AlmostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}
