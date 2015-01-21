package activitystream

import (
	"time"
)

type ObjectType string

type Activity struct {
	Id        string     `bson:"_id" json:"_id,omitempty"`
	Published time.Time  `bson:"published" json:"published,omitempty"`
	Verb      string     `bson:"verb" json:"verb,omitempty"`
	Actor     BaseObject `bson:"actor" json:"actor,omitempty"`
	Object    BaseObject `bson:"object" json:"object,omitempty"`
	Target    BaseObject `bson:"target" json:"target,omitempty"`
	Version   string     `bson:"version" json:"version,omitempty"`
}

func (a *Activity) Score() int {
	return int(MakeTimestamp(a.Published))
}

type Actor BaseObject

type Object BaseObject

type Target BaseObject

type BaseObject struct {
	Id          string            `bson:"id" json:"id,omitempty"`
	URL         string            `bson:"url" json:"url,omitempty"`
	ObjectType  ObjectType        `bson:"objectType" json:"objectType,omitempty"`
	Image       Image             `bson:"image" json:"image,omitempty"`
	DisplayName string            `bson:"displayName" json:"displayName,omitempty"`
	Content     string            `bson:"content" json:"content,omitempty"`
	Metadata    map[string]string `bson:"metadata" json:"metadata,omitempty"`
}

type Image struct {
	URL    string `bson:"url" json:"url,omitempty"`
	Width  int    `bson:"width" json:"width,omitempty"`
	Height int    `bson:"height" json:"height,omitempty"`
}
