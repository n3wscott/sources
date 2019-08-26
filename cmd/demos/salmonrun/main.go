package main

import (
	"context"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/n3wscott/sources/cmd/demos/salmonrun/pkg/controller"
	moron "github.com/spencer-p/moroncloudevents"
)

type Config struct {
	// Serving options
	Port     string `envconfig:"PORT" required:"true"`
	DataPath string `envconfig:"KO_DATA_PATH" default:"./kodata/"`

	// Sourcing options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`

	// The role we are fulfilling: "salmon" or "bear"
	Role string `envconfig:"SALMONRUN_ROLE" required:"true"`
}

func main() {
	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	svr, err := moron.NewServer(&moron.ServerConfig{
		Port:                  env.Port,
		CloudEventReceivePath: "/receive",
		Target:                env.Sink,
	})
	if err != nil {
		log.Fatal("Could not create server: ", err)
	}

	controller.RegisterHandlers(svr, env.Role, env.DataPath)

	svr.HandleCloudEvents(receive)

	log.Fatal(svr.ListenAndServe())
}
