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
