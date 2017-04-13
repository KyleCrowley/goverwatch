package main

import (
	"strconv"
	"strings"
)

// TrimToInt returns an cleaned int given a string.
// Various "cleaning operations" include stripping of whitespace and removal of commas.
func TrimToInt(s string) int {
	// NOTE: String s may contain a comma, so we need to strip out all commas (replace each with empty string).
	s = strings.Replace(s, ",", "", -1)

	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

// TrimToString returns a cleaned string given a string.
// Various "cleaning operations" include stripping of whitespace.
func TrimToString(s string) string {
	return string(strings.TrimSpace(s))
}

// TrimToFloat returns an float given a string.
// Various "cleaning operations" include stripping of whitespace and removal of commas.
func TrimToFloat(s string) float64 {
	// NOTE: String s may contain a comma, so we need to strip out all commas (replace each with empty string).
	s = strings.Replace(s, ",", "", -1)

	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return f
}

// CalculateStars calculates the number of stars the player has according to their (true) level.
// Stars are awarded every 100 levels, but reset every 600 levels, such that 5 stars can be earned per tier:
// Bronze (1-600): 		Star at 101, 201, 301, 401, 501
// Silver (601-1200): 		Star at 601, 701, 801, 901, 1001, 1101
// Gold (1201-1800): 		Star at 1201, 1301, 1401, 1501, 1601, 1701
// Platinum (1801-2400): 	Star at 1801, 1901, 2001, 2101, 2201, 2301
// Above Platinum (2400+):	5 Stars, unchanged regardless of level changes.
// See http://overwatch.wikia.com/wiki/Progression#Lookup_table_and_portrait_border_gallery for a detailed breakdown.
func CalculateStars(level int) int {
	stars := 0

	switch {
	case level > 2400:
		stars = 5
		break
	case level > 1800 && level < 2401:
		stars = CalculateStars(level - 1800)
		break
	case level > 1200 && level < 1801:
		stars = CalculateStars(level - 1200)
		break
	case level > 600 && level < 1201:
		stars = CalculateStars(level - 600)
		break
	// Level is between 1-600, inclusive.
	// Calculation for stars = floor(level/100).
	default:
		stars = int(level / 100)
	}

	return stars
}
