ActivityStream
==============

There are two main parts when implementing activity stream with fan-out on write: Aggregate interested parties of a certain
activity and write to their streams. This project implements the second part efficiently using Redis.

It provides an interface for storing and retrieving streams of Activities. It includes an implementation using Redis and
providing pagination.
An Activity stream (or social feed) is defined as a collection of Activities sorted by time of creation/publication.
The struct Activity provides a format for storing an Activity, it is based on the definition on (activitystrea.ms)[http://activitystrea.ms/].

## Example Architecture
### Requirements

People can follow each other and see their Followings' activities as a homestream and a person's activity on her/his
profile. So we need the following:

- An outbox stream for every person, identified as PERSON_ID-out.
- inbox stream for every person, identified as PERSON_ID.

When an activity occurs:

1. Create a new Activity object.
2. Retrieve list of followers
3. Call activitystream.AddToStream with the activity and the followers inbox-ids.
4. Call activitystream.AddToStream with the activity and the actor's outbox-id.

#### API

In our case I implemented an API service which accepts new activities, aggregates interested parties (followers), stores activities and returns streams.

##### Data returned for an outbox
![data](https://cloud.githubusercontent.com/assets/6203829/5836173/615b9826-a17a-11e4-980b-b2ec98a9d1d5.png)

##### Links returned for an outbox
![links](https://cloud.githubusercontent.com/assets/6203829/5836175/675e71a8-a17a-11e4-9052-0e259691dea3.png)


