package redisstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrisport/go-activitystream/activitystream"
	redis "github.com/garyburd/redigo/redis"
	"reflect"
	"time"
)

const (
	// RedisDefaultProtocol is the default protocol used to connect to Redis, in case no other is specified.
	RedisDefaultProtocol = "tcp"
	// RedisDefaultURL is the default url used to connect to Redis, in case no other is specified.
	RedisDefaultURL = ":6379"

	luaResolveStreamSetAll = `local ids=redis.call("ZREVRANGE",KEYS[1],0,ARGV[1])
	if table.getn(ids)==0 then return {} end
	return redis.call("MGET",unpack(ids))`
	// AFTER:  ZREVRANGEBYSCORE 5444ccbae3c1290013000004-out 1421679584 -inf LIMIT 1 2
	luaResolveStreamSetAfter = `local ids=redis.call("ZREVRANGEBYSCORE",KEYS[1],ARGV[1],"-inf","LIMIT",1,ARGV[2])
	if table.getn(ids)==0 then return {} end
	return redis.call("MGET",unpack(ids))`
	// BEFORE: ZRANGEBYSCORE 5444ccbae3c1290013000004-out 1421679584 +inf LIMIT 1 2
	luaResolveStreamSetBefore = `local ids=redis.call("ZRANGEBYSCORE",KEYS[1],ARGV[1],"+inf","LIMIT",1,ARGV[2])
	if table.getn(ids)==0 then return {} end
	return redis.call("MGET",unpack(ids))`
)

// NewRedisActivityStream returns a new RedisActivityStream, ready to use.
func NewRedisActivityStream(protocol, url string) activitystream.ActivityStream {
	as := RedisActivityStream{
		maxStreamSize: activitystream.DefaultMaxStreamSize,
	}
	as.Init(protocol, url)
	return &as
}

// RedisActivityStream is an implementation of ActivityStream using Redis.
type RedisActivityStream struct {
	pool          *redis.Pool
	maxStreamSize int
}

// SetMaxStreamSize will set the maximum number of elements of a stream to the specified number.
// A negative number means there is no limit, the streams will keep growing.
// Important: Decreasing this number will
// 		1. not affect existing streams unless a new element is added.
// 		2. by adding a new element to an existing stream, the stream will be cut down to the new maximum
func (as *RedisActivityStream) SetMaxStreamSize(maxStreamSize int) {
	if maxStreamSize == 0 {
		maxStreamSize = activitystream.DefaultMaxStreamSize
	}
	as.maxStreamSize = maxStreamSize + 1
}

// Init initializes the RedisActivityStream, it takes exactly two arguments:
//	protocol		the protocol to connect to redis, "tcp" by default
//	url		the url of redis including the port, ":6379" by default
func (as *RedisActivityStream) Init(args ...string) {
	if len(args) != 2 {
		args = []string{RedisDefaultProtocol, RedisDefaultURL}
		fmt.Println("RedisActivityStream:Init(): number of args not equal 2; use default values instead")
	}
	if args[0] == "" {
		args[0] = RedisDefaultProtocol
		fmt.Println("RedisActivityStream:Init(): protocol was empty, use default instead (\"" + RedisDefaultProtocol + "\")")
	}
	if args[1] == "" {
		args[1] = RedisDefaultURL
		fmt.Println("RedisActivityStream:Init(): url was empty, use default instead (\"" + RedisDefaultURL + "\")")
	}

	if as.maxStreamSize == 0 {
		as.maxStreamSize = activitystream.DefaultMaxStreamSize
	}
	values := []string{args[0], args[1]}
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial(values[0], values[1])
		return
	}

	// initialize a new pool
	as.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (as *RedisActivityStream) execute(cmd string, args ...interface{}) (result interface{}, err error) {
	c := as.pool.Get()
	defer c.Close()

	return c.Do(cmd, args...)
}

// GetStream returns an array of Activities belonging to a certain stream. First element is newest.
// The stream is identified by its ID.
// Pagination is provided as follow:
//	size		the size of the page
//	pivotTime		the last received unix time in millisecond, used for identifying page start
//	direction	the direction from pivotTime, the page starts either After the pivot or Before the pivot
func (as *RedisActivityStream) GetStream(streamId string, size int, pivotTime int, afterNotBefore activitystream.Direction) ([]activitystream.Activity, error) {
	var raw interface{}
	var err error
	if pivotTime == 0 {
		raw, err = as.execute("eval", luaResolveStreamSetAll, 1, streamId, size-1)
	} else if afterNotBefore == activitystream.After {
		// AFTER:  ZREVRANGEBYSCORE
		raw, err = as.execute("eval", luaResolveStreamSetAfter, 1, streamId, pivotTime, size)
	} else {
		// BEFORE: ZRANGEBYSCORE
		raw, err = as.execute("eval", luaResolveStreamSetBefore, 1, streamId, pivotTime, size)
	}

	if err != nil {
		return nil, err
	}

	reply, ok := raw.([]interface{})
	if !ok {
		rawType := reflect.TypeOf(raw)
		return nil, errors.New("Redis response was invalid. Response was of type " + rawType.String())
	}
	if afterNotBefore == activitystream.Before {
		// reverse the array since we used original order from Redis for before-request (oldest->newest)
		for i, j := 0, len(reply)-1; i < j; i, j = i+1, j-1 {
			reply[i], reply[j] = reply[j], reply[i]
		}
	}

	activities := make([]activitystream.Activity, 0)
	for i := range reply {
		activity, err := parseActivityFromResponse(reply[i], err)
		if err != nil {
			continue
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// BulkGet returns an array of Activity by their IDs
func (as *RedisActivityStream) BulkGet(ids ...string) ([]activitystream.Activity, error) {
	c := as.pool.Get()
	defer c.Close()

	for i := range ids {
		c.Send("GET", ids[i])
	}
	c.Flush()

	errs := make([]error, 0)
	activities := make([]activitystream.Activity, 0)
	for _ = range ids {
		v, err := c.Receive()

		activity, err := parseActivityFromResponse(v, err)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// Get returns a single Activity by its ID
func (as *RedisActivityStream) Get(id string) (activity activitystream.Activity, err error) {
	resp, err := as.execute("GET", id)
	return parseActivityFromResponse(resp, err)
}

// Store stores a single Activity in the database
// This method is idempotent since the Activity is identified by its ID.
func (as *RedisActivityStream) Store(activity activitystream.Activity) error {
	if activity.Published.Unix() <= 0 {
		activity.Published = time.Now().UTC()
	}
	a, err := json.Marshal(activity)
	if err != nil {
		return errors.New("marshalling Activity failed, " + err.Error())
	}
	_, err = as.execute("SET", activity.Id, a)
	return err
}

// AddToStreams adds a certain activity to one or more streams. The streams are identified by their IDs
// Important: This will also write the activity to database, a call to the method 'Store' would be unnecessary but have no effect.
func (as *RedisActivityStream) AddToStreams(activity activitystream.Activity, streamIds ...string) []error {
	resp, err := as.execute("EXISTS", activity.Id)
	if v, ok := resp.(int64); err != nil || (ok && v == 0) {
		err := as.Store(activity)
		if err != nil {
			return []error{err}
		}
	}

	c := as.pool.Get()
	defer c.Close()
	score := activitystream.MakeTimestamp(activity.Published)
	if score <= 0 {
		activity.Published = time.Now().UTC()
		score = activitystream.MakeTimestamp(time.Now().UTC())
	}

	idHex := activity.Id
	for i := range streamIds {
		c.Send("ZADD", streamIds[i], score, idHex)
		if as.maxStreamSize > 0 {
			c.Send("ZREMRANGEBYRANK", streamIds[i], 0, -as.maxStreamSize)
		}
	}
	c.Flush()

	k := len(streamIds)
	if as.maxStreamSize > 0 {
		k = len(streamIds) * 2
	}

	errs := make([]error, 0)
	for ; k > 0; k-- {
		if _, err := c.Receive(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func parseActivityFromResponse(resp interface{}, respErr error) (activity activitystream.Activity, err error) {
	if respErr != nil {
		return activity, respErr
	}
	if resp == nil {
		err = redis.ErrNil
		return
	}

	val, ok := resp.([]byte)
	if !ok {
		err = errors.New("item from redis is not a byte array")
		return
	}

	err = json.Unmarshal(val, &activity)
	if err != nil {
		err = errors.New("unmarshall Activity failed, " + err.Error())
		return
	} else if activity.Id == "" {
		err = errors.New("object was not valid activity")
		return
	}
	return
}
