ActivityStream
==============

Status:
[![Build Status](https://drone.io/github.com/chrisport/go-activitystream/status.png)](https://drone.io/github.com/chrisport/go-activitystream/latest)
[![Coverage Status](https://coveralls.io/repos/chrisport/go-activitystream/badge.svg)](https://coveralls.io/r/chrisport/go-activitystream)


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
![data_compact](https://cloud.githubusercontent.com/assets/6203829/5836435/6abf546c-a17e-11e4-929e-3aeb399b7478.png)

##### Links returned for an outbox
![links](https://cloud.githubusercontent.com/assets/6203829/5836175/675e71a8-a17a-11e4-9052-0e259691dea3.png)

## Contribution

Suggestions and Bug reports can be made through Github issues.
Contributions are welcomed, there is currently no need to open an issue for it, but please follow the code style, including descriptive tests with [GoConvey](http://goconvey.co/).

## License

Licensed under [Apache 2.0](LICENSE).
