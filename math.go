package main

import (
	"math"
)

// compute absolute value (modulus) of complex number
func size(x complex128) float64 {
	// return math.Hypot(real(x), imag(x)) // slower
	return math.Sqrt(real(x)*real(x) + imag(x)*imag(x))
}

// determines whether the complex number is in the manderbrot set
// Z = Z^2+C
func isInSet(c complex128, maxSteps int) (isIn bool, steps float64) {
	i := 1
	z := 0 + 0i
	for i < maxSteps {
		z = z*z + c
		if real(z)*real(z)+imag(z)*imag(z) > 4 {
			steps = float64(i) + 1 - math.Log(math.Log2(size(z))) // smoothen the steps
			return
		}
		i++
	}
	isIn = true // we assume it does not escape the circle afther this point
	return
}

// n must be positive odd integer
// 0 1 2 3 4 5 6 7 =>
// 4 3 5 2 6 1 7 0
func zigZagIndexes(n int) []int {
	list := make([]int, n)

	for i, r, l := 0, n/2, n/2-1; i < n; i, r, l = i+2, r+1, l-1 {
		list[i] = l
		list[i+1] = r
	}
	return list
}

// return n multiplied by sqrt2 if sign is +1
// or n divided by sqrt2 if sign is -1
// powers of 2 are rounded exactly
func nextSqrt(n float64, sign int) float64 {
	lvl := int(math.Round(math.Log2(n) * 2))
	lvl += sign

	return math.Pow(2, float64(lvl)/2)
}

func valueToColor(v float64, coloring int, maxSteps int) (byte, byte, byte) {
	li := float64(16) // lightness 0-255
	st := float64(1)  // steps/circle granulity
	shade := byte(0)
	switch coloring {
	case 0:
		shade = byte(li + math.Log(v*st)/math.Log(float64(maxSteps))*float64((256-li)))
	case 1:
		shade = byte(li + math.Mod(v*st, 256-li))
	case 2:
		shade = byte(li + (v*st)/float64(maxSteps)*(256-li))
	case 3:
		shade = byte(li + (v*st)/float64(maxSteps)*(256-li))
		if shade > byte(li)+byte(256-int(li))/2 { // flip gradient direction
			shade = byte(li) + byte(256-int(li)) - shade
		}
	}
	return shade, shade, shade
}
