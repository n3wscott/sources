package controller

import (
	"context"
	"log"
	"net/http"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}

// A ConnectionReceiver is used to process messages on a websocket.
type ConnectionReceiver func(player Player, conn *websocket.Conn)

// A ConnectionSender sends messages on a websocket.
type ConnectionSender func(msg *Message)

func makeWebSocketHandle(callback ConnectionReceiver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("websocket")

		name := r.FormValue("name")
		if name == "" {
			// Form validation is for serious applications, not demos
			name = "Steve"
		}

		log.Printf("Websocket request with name %q\n", name)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection: ", err)
			return
		}

		player := Player{
			Name: name,
			UUID: uuid.New().String(),
		}

		makeConnSender(player, conn)
		go callback(player, conn)
	}
}

func makeSalmonWSReceiver(client cloudevents.Client) ConnectionReceiver {
	return func(player Player, conn *websocket.Conn) {
		// TODO it would be nice to have a way to cancel this
		for {
			mtype, _, err := conn.ReadMessage()
			if err != nil {
				if closeerror(err) {
					return
				}
				log.Println("conn read err: ", err)
				continue
			} else if mtype != websocket.TextMessage {
				log.Println("expected text message format, got something else")
				continue
			}

			// TODO it would be nice to validate the contents of the message.
			// But it doesn't matter. The salmon only has one thing to say: "jump".
			log.Println("Got a websocket message")

			event := cloudevents.NewEvent()
			event.SetSource(EVENT_SOURCE)
			event.SetType(SALMON_EVENT_TYPE)
			if err := event.SetData(&Message{
				Type: "jump",
				From: player,
			}); err != nil {
				log.Println("Failed to set cloud event data: ", err)
				continue
			}

			client.Send(context.Background(), event)
		}
	}
}

func makeBearWSReceiver(client cloudevents.Client) ConnectionReceiver {
	return func(player Player, conn *websocket.Conn) {
		timeoutchan := make(chan Message)
		timeoutchans[player.Key()] = timeoutchan

		msgchan := make(chan Message)
		stopchan := make(chan struct{})
		seen := make(map[string]struct{})

		go func() {
			var msg Message
			for {
				if err := conn.ReadJSON(&msg); err != nil {
					if closeerror(err) {
						close(msgchan)
						stopchan <- struct{}{}
						return
					}
					log.Println("failed to read json msg: ", err)
					continue
				}
				log.Println("Received a websocket message")
				msgchan <- msg
			}

		}()

		for {
			var msg Message
			select {
			case msg = <-timeoutchan:
				seen[msg.Nonce] = struct{}{}
				msg.Type = "noteat"
				log.Println("Bear missed the fish!")
			case msg = <-msgchan:
				if _, ok := seen[msg.Nonce]; ok {
					continue
				}
				msg.Type = "eat"
				log.Println("Bear got the fish!")
			case <-stopchan:
				return
			}

			// Bear eat messages have who they are going to (the person that gets eaten).
			// They don't know themselves though; we have to add that.
			msg.From = player

			event := cloudevents.NewEvent()
			event.SetSource(EVENT_SOURCE)
			event.SetType(BEAR_EVENT_TYPE)
			if err := event.SetData(&msg); err != nil {
				log.Println("Failed to set cloud event data: ", err)
				continue
			}

			client.Send(context.Background(), event)
		}
	}
}

func makeConnSender(player Player, conn *websocket.Conn) ConnectionSender {
	connclosed := false
	var mtx sync.Mutex

	f := func(msg *Message) {
		if connclosed == true {
			// No client to send to, drop msg on the floor
			return
		}

		// Only one writer may write to the websocket.
		mtx.Lock()
		defer mtx.Unlock()
		if err := conn.WriteJSON(msg); err != nil {
			if closeerror(err) {
				connclosed = true
				delete(conns, player.Key())
				return
			}
			log.Println("conn write json failed: ", err)
		}
	}

	conns[player.Key()] = f
	return f
}

func closeerror(err error) bool {
	return websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
		websocket.CloseNoStatusReceived,
	)
}
