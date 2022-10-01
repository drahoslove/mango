package main

import (
	"math"
	"strconv"
)

func toHumNum(n float64) string {
	unit := ""
	for _, u := range []string{"k", "M", "G", "T", "P"} {
		if n >= 1<<10 {
			n /= 1 << 10
			unit = u
		}
	}
	_n := math.Floor(n)
	if _n == n {
		return strconv.FormatInt(int64(n), 10) + unit
	}
	return strconv.FormatFloat(n, 'f', 1, 64) + unit
}
