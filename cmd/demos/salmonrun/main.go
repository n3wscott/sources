package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/cmd/demos/salmonrun/pkg/controller"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
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
		Port:                  conf.Port,
		CloudEventReceivePath: "/receive",
		CloudEventTargets:     []string{conf.Sink},
	})
	if err != nil {
		log.Fatal("Could not create server: ", err)
	}

	err = controller.RegisterHandlers(svr, conf.Role, conf.DataPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(svr.ListenAndServe())
}
