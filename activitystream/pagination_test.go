package activitystream

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	testing "testing"
	"time"
)

func TestNextAndPrevTokenGeneration(t *testing.T) {
	activities := make([]Activity, 3)
	activities[2] = createTestActivity()
	time.Sleep(100)
	activities[1] = createTestActivity()
	time.Sleep(100)
	activities[0] = createTestActivity()

	timeStampNewest := MakeTimestamp(activities[0].Published)
	timeStampOldest := MakeTimestamp(activities[2].Published)
	afterNotBefore := After

	Convey("Subject: Test createLinks", t, func() {
		Convey("When size of result is equals desired size", func() {
			size := 3
			Convey("It should return correct next and prev", func() {
				correctNext := fmt.Sprintf("?s=%d&after=%d", size, timeStampOldest)
				correctPrev := fmt.Sprintf("?s=%d&before=%d", size, timeStampNewest)

				prev, next := CreateTokens(size, afterNotBefore, activities)

				So(next, ShouldEqual, correctNext)
				So(prev, ShouldEqual, correctPrev)
			})
		})
		Convey("When size of result is smaller then desired size and we wanted after pivot", func() {
			size := 4
			afterNotBefore := After
			correctPrev := fmt.Sprintf("?s=%d&before=%d", size, timeStampNewest)

			prev, next := CreateTokens(size, afterNotBefore, activities)

			Convey("It should return correct prev and no next", func() {
				So(next, ShouldBeEmpty)
				So(prev, ShouldEqual, correctPrev)
			})
		})

		Convey("When size of result is smaller then desired size and we wanted before pivot", func() {
			size := 4
			afterNotBefore := Before
			correctNext := fmt.Sprintf("?s=%d&after=%d", size, timeStampOldest)
			prev, next := CreateTokens(size, afterNotBefore, activities)

			Convey("It should return correct next and no prev", func() {
				So(prev, ShouldBeEmpty)
				So(next, ShouldEqual, correctNext)
			})
		})

		Convey("When empty array of activites", func() {
			size := 4
			afterNotBefore := Before
			prev, next := CreateTokens(size, afterNotBefore, []Activity{})

			Convey("It should return no next and no prev", func() {
				So(next, ShouldBeEmpty)
				So(prev, ShouldBeEmpty)
			})
		})
	})
}

func createTestActivity() Activity {
	var a Activity
	a.Id = bson.NewObjectId().Hex()
	a.Published = time.Now().UTC()
	a.Verb = "JOIN"
	a.Actor = BaseObject{}
	a.Actor.Id = "ACTOR_ID"
	a.Actor.ObjectType = "Profile"

	a.Object = BaseObject{}
	a.Object.ObjectType = "Community"
	a.Object.Id = "COMMUNITY_ID"
	return a
}
