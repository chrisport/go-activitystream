package redisstream

import (
	"github.com/chrisport/go-activitystream/activitystream"
	redis "github.com/garyburd/redigo/redis"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	testing "testing"
	"time"
)

const (
	skipIntegrationTests = false
	address              = ":6379"
	protocol             = "tcp"
)

func TestCreateRedisActivitystream(t *testing.T) {
	if skipIntegrationTests {
		return
	}

	testActivity := createTestActivity()

	Convey("Subject: Test Creating new RedisActivitystream", t, func() {
		Convey("When acitivitystream is retrieved through NewRedisActivitStream method", func() {
			asUnderTest := NewRedisActivityStream(protocol, address)

			Convey("It should be ready to Store and Get an activity", func() {
				defer removeFromRedis(testActivity.Id)
				err := asUnderTest.Store(testActivity)
				So(err, ShouldBeNil)

				res, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(res, testActivity), ShouldBeTrue)
			})
		})
		Convey("When acitivitystream is created and Init is called", func() {
			asUnderTest := RedisActivityStream{}
			asUnderTest.Init(protocol, address)

			Convey("It should be ready to Store and Get an activity", func() {
				defer removeFromRedis(testActivity.Id)
				err := asUnderTest.Store(testActivity)
				So(err, ShouldBeNil)

				res, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(res, testActivity), ShouldBeTrue)
			})
		})
		Convey("When acitivitystream is created with missing arguments", func() {
			asUnderTest := RedisActivityStream{}
			asUnderTest.Init("")

			Convey("It should use default values and be ready to Store and Get an activity", func() {
				defer removeFromRedis(testActivity.Id)
				err := asUnderTest.Store(testActivity)
				So(err, ShouldBeNil)

				res, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(res, testActivity), ShouldBeTrue)
			})
		})
		Convey("When acitivitystream is created with empty arguments", func() {
			asUnderTest := RedisActivityStream{}
			asUnderTest.Init("", "")

			Convey("It should use default values and be ready to Store and Get an activity", func() {
				defer removeFromRedis(testActivity.Id)
				err := asUnderTest.Store(testActivity)
				So(err, ShouldBeNil)

				res, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(res, testActivity), ShouldBeTrue)
			})
		})
	})
}

func TestStoreAndGet(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testActivity := createTestActivity()

	Convey("Subject: Test Store and Get Activity", t, func() {
		Convey("When activity is written", func() {
			err := asUnderTest.Store(testActivity)
			So(err, ShouldBeNil)
			defer removeFromRedis(testActivity.Id)

			Convey("It should be available through Get", func() {
				res, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(res, testActivity), ShouldBeTrue)

			})
			Convey("When activity without Published timestamp is written", func() {
				originalTimestamp := testActivity.Published
				defer func() { testActivity.Published = originalTimestamp }()
				defer removeFromRedis(testActivity.Id)
				testActivity.Published = time.Time{}

				err := asUnderTest.Store(testActivity)
				So(err, ShouldBeNil)

				Convey("It should use current time as publish date", func() {
					res, err := asUnderTest.Get(testActivity.Id)
					So(err, ShouldBeNil)
					So(time.Now().UnixNano()-res.Published.UnixNano(), ShouldBeBetweenOrEqual, 10*time.Nanosecond, 5*time.Second)
				})
			})

			Convey("When String is written to a key and Get called on this key", func() {
				_, err := asUnderTest.execute("SET", "SOME_KEY", "NOT_AN_ACTIVITY")
				So(err, ShouldBeNil)
				defer removeFromRedis("SOME_KEY")

				Convey("It should return error", func() {
					_, err := asUnderTest.Get("SOME_KEY")
					So(err, ShouldNotBeNil)

				})
			})

			Convey("When Get called on inexistent key", func() {
				Convey("It should return error", func() {
					_, err := asUnderTest.Get("SOME_INEXISTENT_KEY")
					So(err, ShouldEqual, redis.ErrNil)
				})
			})
		})
	})
}

func TestBulkGet(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testActivity := []activitystream.Activity{
		createTestActivity(),
		createTestActivity(),
		createTestActivity(),
		createTestActivity(),
		createTestActivity(),
	}

	Convey("Subject: Test Store and BulkGet ", t, func() {
		Convey("When ID is written to streams", func() {
			for i := range testActivity {
				err := asUnderTest.Store(testActivity[i])
				So(err, ShouldBeEmpty)
			}
			ids := make([]string, 0)
			for i := range testActivity {
				ids = append(ids, testActivity[i].Id)
			}
			defer removeFromRedis(ids...)

			Convey("It should be available through Get on all these streams", func() {
				activities, err := asUnderTest.BulkGet(ids...)
				So(err, ShouldBeNil)

				for i := range testActivity {
					So(activitiesAreEqual(testActivity[i], activities[i]), ShouldBeTrue)
				}

			})
		})
	})
}

func TestStoreAndGetIDsFromStream(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testActivity := createTestActivity()
	testIDs := []string{"ID1", "ID2", "ID3", "ID4", "ID5"}
	removeFromRedis(testActivity.Id)
	Convey("Subject: Test AddToStreams and GetStream ", t, func() {
		//Precondition
		_, err := asUnderTest.Get(testActivity.Id)
		So(err, ShouldEqual, redis.ErrNil)

		Convey("When activity is written to streams", func() {
			errs := asUnderTest.AddToStreams(testActivity, testIDs...)
			So(errs, ShouldBeEmpty)
			defer removeFromRedis(testActivity.Id)
			defer removeFromRedis(testIDs...)

			Convey("It should be available through Get on all these streams", func() {
				for _, id := range testIDs {
					stream, err := asUnderTest.GetStream(id, 0, 0, activitystream.After)
					So(err, ShouldBeNil)
					So(stream[0].Id, ShouldEqual, testActivity.Id)
				}
			})
			Convey("It should automatically add Activity if not exist", func() {
				insertedActivity, err := asUnderTest.Get(testActivity.Id)
				So(err, ShouldBeNil)
				So(activitiesAreEqual(insertedActivity, testActivity), ShouldBeTrue)
			})

		})
	})
}

func TestAddToStreams(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testActivity := createTestActivity()
	testStreamID := "TEST_STREAM_ID"

	Convey("Subject: Test AddToStream edge-cases ", t, func() {
		Convey("When same ID is written to a stream twice", func() {
			defer removeFromRedis(testStreamID)
			//precondition
			removeFromRedis(testStreamID)
			stream, err := asUnderTest.GetStream(testStreamID, 99, 0, activitystream.After)
			So(err, ShouldBeEmpty)
			So(len(stream), ShouldEqual, 0)

			//write to stream twice
			errs := asUnderTest.AddToStreams(testActivity, testStreamID)
			So(errs, ShouldBeEmpty)
			errs = asUnderTest.AddToStreams(testActivity, testStreamID)
			So(errs, ShouldBeEmpty)

			Convey("It should be there just once", func() {
				//ensure stream has exactly one element
				stream, err = asUnderTest.GetStream(testStreamID, 99, 0, activitystream.After)
				So(err, ShouldBeEmpty)
				So(len(stream), ShouldEqual, 1)
				So(stream[0].Id, ShouldEqual, testActivity.Id)

			})
		})
		Convey("When 150 IDs are written to a stream and Max stream size has been set to 40", func() {
			asUnderTest.SetMaxStreamSize(40)
			defer asUnderTest.SetMaxStreamSize(50)
			Convey("It should trim the stream to 100 items", func() {
				defer removeFromRedis(testStreamID)

				//ensure stream is empty
				removeFromRedis(testStreamID)
				stream, err := asUnderTest.GetStream(testStreamID, 100, 0, activitystream.After)
				So(err, ShouldBeEmpty)
				So(len(stream), ShouldEqual, 0)

				ids := make([]string, 0)
				//write to stream 150 times
				for i := 0; i < 150; i++ {
					activity := createTestActivity()
					ids = append(ids, activity.Id)
					errs := asUnderTest.AddToStreams(activity, testStreamID)
					So(errs, ShouldBeEmpty)
				}

				//ensure stream has exactly one element
				stream, err = asUnderTest.GetStream(testStreamID, 99, 0, activitystream.After)
				So(err, ShouldBeEmpty)
				So(len(stream), ShouldEqual, 40)
				idsFromStream := make([]string, 40)
				for k := 0; k < 40; k++ {
					idsFromStream[k] = stream[k].Id
				}

				for k := 0; k < 40; k++ {
					So(idsFromStream, ShouldContain, ids[k+110])
				}

			})
		})
	})
}

func TestGetStream(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testStreamID := "STREAM_ID"
	testActivity1 := createTestActivity()
	testActivity2 := createTestActivity()
	testActivity3 := createTestActivity()
	defer removeFromRedis(testStreamID, testActivity1.Id, testActivity2.Id, testActivity3.Id)

	Convey("Subject: Test Store and Get complete stream", t, func() {
		//PRECONDITION: ensure stream is empty
		removeFromRedis(testStreamID, testActivity1.Id, testActivity2.Id, testActivity3.Id)
		stream, err := asUnderTest.GetStream(testStreamID, 99, 0, activitystream.After)
		So(err, ShouldBeEmpty)
		So(len(stream), ShouldEqual, 0)

		//SETUP
		err = asUnderTest.Store(testActivity1)
		So(err, ShouldBeNil)
		err = asUnderTest.Store(testActivity2)
		So(err, ShouldBeNil)
		err = asUnderTest.Store(testActivity3)
		So(err, ShouldBeNil)

		errs := asUnderTest.AddToStreams(testActivity1, testStreamID)
		So(errs, ShouldBeEmpty)
		errs = asUnderTest.AddToStreams(testActivity2, testStreamID)
		So(errs, ShouldBeEmpty)
		errs = asUnderTest.AddToStreams(testActivity3, testStreamID)
		So(errs, ShouldBeEmpty)

		Convey("When 3 activities are written to test stream", func() {

			Convey("Last insterted activity should be returned when limit is 1 and afterID empty", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 1, 0, activitystream.After)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 1)
				So(activitiesAreEqual(returnedActivities[0], testActivity3), ShouldBeTrue)
			})

			Convey("oldest and second oldest activities should be returned when limit is 5 and afterID is newest (last inserted one)", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 2, testActivity3.Score(), activitystream.After)
				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 2)
				So(activitiesAreEqual(returnedActivities[0], testActivity2), ShouldBeTrue)
				So(activitiesAreEqual(returnedActivities[1], testActivity1), ShouldBeTrue)
			})

			Convey("second newest activity should be returned when limit is 1 and beforeID is third newest", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 1, testActivity1.Score(), activitystream.Before)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 1)
				So(activitiesAreEqual(returnedActivities[0], testActivity2), ShouldBeTrue)
			})
		})
	})
}

func TestGetStreamEdgeCases(t *testing.T) {
	if skipIntegrationTests {
		return
	}
	asUnderTest := RedisActivityStream{}
	asUnderTest.Init(protocol, address)

	testStreamID := "STREAM_ID"
	testActivity1 := createTestActivity()
	defer removeFromRedis(testStreamID, testActivity1.Id)

	Convey("Subject: Test Store and Get complete stream edge cases", t, func() {
		//PRECONDITION: ensure stream is empty
		removeFromRedis(testStreamID, testActivity1.Id)
		stream, err := asUnderTest.GetStream(testStreamID, 99, 0, activitystream.After)
		So(err, ShouldBeEmpty)
		So(len(stream), ShouldEqual, 0)

		Convey("When 0 activities are written to test stream", func() {
			Convey("It should return empty stream when after ID with a random ID", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 5, testActivity1.Score(), activitystream.After)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 0)
			})
			Convey("It should return empty stream when before ID with a random ID", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 5, testActivity1.Score(), activitystream.Before)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 0)
			})
		})

		Convey("When 1 activity is written to test stream", func() {
			//SETUP
			err = asUnderTest.Store(testActivity1)
			So(err, ShouldBeNil)

			errs := asUnderTest.AddToStreams(testActivity1, testStreamID)
			So(errs, ShouldBeEmpty)

			Convey("empty stream should be returned when we want after ID given the one existing ID", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 5, testActivity1.Score(), activitystream.After)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 0)
			})
			Convey("empty activity should be returned when we want before ID given the one existing ID", func() {
				returnedActivities, err := asUnderTest.GetStream(testStreamID, 5, testActivity1.Score(), activitystream.Before)

				So(err, ShouldBeNil)
				So(len(returnedActivities), ShouldEqual, 0)
			})
		})
	})
}

// ************* HELPER METHODS *************
func createTestActivity() activitystream.Activity {
	var a activitystream.Activity
	a.Id = bson.NewObjectId().Hex()
	a.Published = time.Now().UTC()
	a.Verb = "SOME_VERB_LIKE_CREATE"
	a.Actor = activitystream.BaseObject{}
	a.Actor.Id = "ACTOR_ID"
	a.Actor.ObjectType = "SOME_TYPE_LIKE_PERSON"

	a.Object = activitystream.BaseObject{}
	a.Object.ObjectType = "SOME_TYPE_LIKE_GROUP"
	a.Object.Id = "COMMUNITY_ID"
	return a
}

func activitiesAreEqual(activityA, activityB activitystream.Activity) bool {
	return activityA.Id == activityB.Id &&
		activityA.Verb == activityB.Verb &&
		activityA.Actor.Id == activityB.Actor.Id &&
		activityA.Actor.ObjectType == activityB.Actor.ObjectType &&
		activityA.Object.Id == activityB.Object.Id &&
		activityA.Object.ObjectType == activityB.Object.ObjectType
}

func removeFromRedis(ids ...string) {
	c, err := redis.Dial(protocol, address)
	if err != nil {
		panic(err)
	}

	defer c.Close()

	for _, id := range ids {
		c.Send("DEL", id)
	}
	c.Flush()
}
