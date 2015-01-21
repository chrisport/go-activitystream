package activitystream

import (
	"strconv"
	"time"
)

// MakeTimestamp returns the given time as unix milliseconds
func MakeTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// CreateTokens generates and returns previous and next token from an array of activities and pagination information
// size	the size of the page
// direction	theDirection of the previous request, this is needed for determining first and last page
// activities	the last result
func CreateTokens(size int, direction Direction, activities []Activity) (prev, next string) {
	leng := len(activities)
	if leng == 0 {
		return
	}
	lastPivot := strconv.Itoa(int(MakeTimestamp(activities[leng-1].Published)))
	firstPivot := strconv.Itoa(int(MakeTimestamp(activities[0].Published)))
	s := strconv.Itoa(size)

	if direction == After || leng >= size {
		prev = "?s=" + s + "&before=" + firstPivot
	}

	if direction == Before || leng >= size {
		next = "?s=" + s + "&after=" + lastPivot
	}

	return
}
