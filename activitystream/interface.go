// Package activitystream provides an interface to implement an activitystream.
// Further it contains a default implementation using Redis.
//
// Definition ActivityStream
// 		An ActivityStream is a list of Activities sorted by time of insertion (LIFO)
//
// By this definition an ActivityStream is a list, not a set. Therefore elements that are inserted multiple times
// will also appear multiple times in the stream.
package activitystream

import (
	"github.com/garyburd/redigo/redis"
)

// DefaultMaxStreamSize is the number of elements a stream can store by default.
// This number can be adjusted on the ActivityStream using its method SetMaxStreamSize.
const DefaultMaxStreamSize = 50

// Direction represents the direction a pagination token is going
type Direction bool

const (
	// After Direction indicates in pagination that next page comes After a certain element
	After Direction = true
	// After Direction indicates in pagination that next page comes Before a certain element
	Before Direction = false
)

var ErrEmpty = redis.ErrNil

// ActivityStream interface defines functionality to implement an activity stream. An activity can be stored and added
// to a stream. A stream is always sorted with the newest (last insterted) element on top.
type ActivityStream interface {
	// Init initializes the ActivityStream, arguments are defined by specific implementation
	Init(args ...string)

	// SetMaxStreamSize will set the maximum number of elements of a stream to the specified number.
	// A negative number means there is no limit, the streams will keep growing.
	// Important: Decreasing this number will
	// 		1. not affect existing streams unless a new element is added.
	// 		2. by adding a new element to an existing stream, the stream will be cut down to the new maximum
	SetMaxStreamSize(maxStreamSize int)

	// Get returns a single Activity by its ID
	Get(id string) (activity Activity, err error)

	// BulkGet returns an array of Activity by their IDs
	BulkGet(id ...string) ([]Activity, error)

	// Store stores a single Activity in the database
	// This method is idempotent since the Activity is identified by its ID.
	Store(activity Activity) error

	// GetStream returns an array of Activity belonging to a certain stream. First element is last inserted.
	// The stream is identified by its ID.
	// Pagination is provided as follow:
	//	limit		the size of the page
	//	pivotID		the last received ID, this element will not be included in the result
	//	direction	the direction from pivotID, the page starts either After the pivot or Before the pivot
	GetStream(streamId string, limit int, pivotID int, direction Direction) ([]Activity, error)

	// AddToStreams adds a certain activity to one or more streams. The streams are identified by their IDs
	// Important: This will also write the activity to database, a call to the method 'Store' would be duplicate
	AddToStreams(activity Activity, streamIds ...string) []error
}
