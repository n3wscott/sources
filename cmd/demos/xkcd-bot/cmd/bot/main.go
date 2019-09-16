package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port    string `envconfig:"PORT" default:"8080"`
	Webhook string `envconfig:"GCHAT_WEBHOOK" required:"true"`
}

func main() {
	var env Config
	envconfig.MustProcess("", &env)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read POST data assuming it is a JSON payload of XKCD comic data.
		log.Printf("Received POST with %d bytes of data\n", r.ContentLength)
		var comic XkcdComic
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&comic); err != nil {
			http.Error(w, "Not a valid XKCD comic: "+err.Error(), http.StatusBadRequest)
			log.Println("Not a valid comic:", err)
			return
		}

		// Dump it to GChat in a nice format
		resp, err := http.Post(env.Webhook, "text/json", bytes.NewReader(comic.AsGChatMessage()))
		if err != nil {
			http.Error(w, "Failed to send to webhook", http.StatusInternalServerError)
			log.Println("Failed to send to webhook:", err)
			return
		}

		log.Printf("%s from GChat\n", resp.Status)

		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println("Failed to copy response from GChat:", err)
		}
	})

	log.Printf("Delivering messages to %q\n", env.Webhook)
	log.Println("Starting on port", env.Port)
	http.ListenAndServe(":"+env.Port, nil)
}
