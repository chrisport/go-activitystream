ActivityStream
==============

This project helps implementing an activitystream (social feed) following the **fan-out on write** approach:

1. [on Write] Aggregate interest**ed** parties
2. [on Write] Write/store to their streams
3. [on Read] Retrieve stream

In opposite to **fan-in on read** which would consist of:

1. [on Write] Write/store activities
2. [on Read] Aggregate interest**ing** parties
3. [on Read] Read from their streams (and aggregate them)

It realises the second and third part of a fan-out on write efficiently using Redis. It provides an interface which can be used to replace Redis with other databases/storages (Groupcache would be interesting! I may play with it in future).
The project also defines a way of pagination following Facebook's approach, as well as a format for storing an Activity, which is based on the definition on [activitystrea.ms](http://activitystrea.ms/). The Redis implementation stores activities just once and writes their ID to the specified streams.

## Complete Example Architecture
### Requirements

Assuming in a system people can follow each other, see their Followings' activities as a "homestream" and a person's activity on her/his
profile. So we need the following:

- An **outbox stream** for every person, identified as "PERSON_ID-out", which will be shown on the person's profile.
- An **inbox stream** for every person, identified as "PERSON_ID", which will be shown as homestream of followers.

When an activity occurs:

1. Create a new Activity object.
2. Retrieve list of followers
3. Call activitystream.AddToStream with the activity using the followers inbox-ids + the actor's outbox-id.

#### API

In our case I implemented an API service which accepts new activities, aggregates interested parties (followers), stores activities and returns streams.

##### Data returned for an outbox
![data](https://cloud.githubusercontent.com/assets/6203829/5836173/615b9826-a17a-11e4-980b-b2ec98a9d1d5.png)

##### Links returned for an outbox
![links](https://cloud.githubusercontent.com/assets/6203829/5836175/675e71a8-a17a-11e4-9052-0e259691dea3.png)


