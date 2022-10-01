package main

import (
	"fmt"
	"math"
	"testing"
)

func TestSize(t *testing.T) {
	fmt.Println(size(0 + 0i))
	fmt.Println(size(1 + 1i))

	if size(2+2i) < 2 {
		t.Fatalf("Size of %v shoudl be over 2", 2+2i)
	}
	if size(0+0i) != 0 {
		t.Fatalf("Size of %v shoudl be 0", 0+0i)
	}
}

func TestZigZagList(t *testing.T) {

	list := zigZagIndexes(8)
	expected := []int{3, 4, 2, 5, 1, 6, 0, 7}

	for i, v := range list {
		if expected[i] != v {
			t.Fatalf("zigZagList of %v shoudl be %v not %v", 8, expected, list)
		}
	}
}

func TestNextSqrt(t *testing.T) {

	if nextSqrt(1, 1) != math.Sqrt2 {
		t.Fatalf("nextSqrt times 2 of %v shoudl be %v not %v",
			1, math.Sqrt2, nextSqrt(1, 1))
	}

	zoom := float64(1)
	next2 := nextSqrt(nextSqrt(zoom, 1), 1)

	if next2 != 2 {
		t.Fatalf("nextSqrt times 2 of %v shoudl be %v not %v",
			zoom, 2, next2)
	}

}

func TestToHumNum(t *testing.T) {
	table := [](struct {
		input    float64
		expected string
	}){
		{1000, "1000"},
		{1023, "1023"},
		{1024, "1k"},
		{2200, "2k"},
		{23650, "23k"},
		{42 * 1 << 10 * 1 << 10, "42M"},
	}

	for _, line := range table {
		result := toHumNum(line.input)
		if result != line.expected {
			t.Fatalf("toHumNum of %v should be %v but is %v",
				line.input, line.expected, result)
		}
	}
}
