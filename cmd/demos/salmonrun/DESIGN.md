## Abstract

I am going to write a game (**salmonrun**) where players will play as either a
salmon jumping up the river or a bear trying to catch salmon. The salmon get
points by not getting eaten; the bears get points by eating salmon.

## Motivation

The ServiceSource resource is unique in that it can function as both a source
and an addresabletype. I want to demonstrate a ServiceSource being used as both
at once.

A ServiceSource cannot have itself as a sink. It needs to resolve the sink
before it can create the KService, and the sink cannot be resolved before the
KService is created, which is a dependency loop. For that matter, any closed
loop of addressables will be deadlocked.

TODO(spencer-p): If a ServiceSource was capable of creating the Route before
launching any containers, we could resolve this problem.

To have two ServiceSource communicate with each other, I will set up the graph
as follows:

 - ServiceSource A -> Default broker
 - ServiceSource B -> Default broker
 - Trigger filtering on A -> B
 - Trigger filtering on B -> A

This will prevent a deadlock.

## User Experience

A user will either browse to `http://salmon.default.dev.example.com/` or
`http://bear.default.dev.example.com/`. They will be prompted to enter a
username, and then they will enter the game. The salmon will have a button to
press to jump (perhaps the space bar?). If time permits, they will see a little
salmon jump. If a bear somewhere eats them, they do not get a point, if the bear
does not eat them, they get a point.

The bear will wait for a salmon to appear (first a brief button, then an actual
salmon jumping). If they click on the salmon, they get a point.

Both ends will display the name of the user that ate them/they ate.
 - You ate foobar!
 - Unfortunately, foobar escaped you.
 - You were eaten by foobar!
 - Luckily, foobar did not eat you.

## Frontend design

The root URL will be a simple form displaying either a salmon or a bear and
prompting for a name. This will send a GET to `/game` or something like that
with the username in the query parameters.

The `/game` URL will open up a websocket. The websocket will likely be opened
with the username as well. After this, the username can be thrown out.

On the salmon side: When the player jumps, a web socket message will be sent
("jump" or something simple). Some variation of a jumping animation will play.
The websocket will reply with a message stating whether or not they were eaten,
along with the name of the player they were matched with. A status will be
printed accordingly.

The bear side is similar, except that events will be triggered by an incoming
message. The message will have a name and a timeout for how long they have to
eat the salmon. If they eat the salmon, they must send a message saying so. The
backend will assume that the salmon was not eaten if it does not receive a
message in time.

## Backend design

The root will serve a simple form for the username. Usernames will be appended
with a UUID so that name collisions do not occur.

The game path will serve the html/scripts described above.

There will be some `/websocket` path that will upgrade to a websocket. When a
websocket is created, it will be placed in a pool of active connections. The
stored websocket will retain details about the user (perhaps wrapped in a
struct, or in a closure). Messages will always store the username and UUID. The
pool of connections will be a map (map[string]connection) where the keys are the
cat of username and UUID.

For the salmon: When a jump message is received, create a cloud event with the
salmon type and send it off. Contains username, etc. We will get a response
within the timeout that they were either eaten or not, which will get sent back
via websocket. The response will NOT be a response on the same connection. It
will be a brand new connection that we process on the receive handler.

For the bear: Websocket incoming requests will be the user clicking on a salmon.
This will unblock the jump cloud event to send a cloud event response.

There will be a `/receive` path for CloudEvents.

The salmon will get eat/not eaten responses:
``go
func receive() {
  uname, uuid := from the event
  conn := connectionmap[uname+uuid]
  conn.Send(event data)
}
```

The bear will receive a jump method and start a timeout and wait for a eat on
the websocket. If the user eats the salmon, it will respond as such; if they
don't do it in time, we send the no eat. The receive will generally look like
this:
```go
func receive() {
  // TODO Does each receive get its own goroutine or are these blocking?
  // If these are blocking, this should start a goroutine.
  timeout := make some timeout
  eaten := create some channel to receive from the websocket/look up websocket
  select {
    timeout:
      too bad, respond with no eat
    eaten:
      tell em they got eat!
  }
}
```

