/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
		CloudEventReceivePath: "/",
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
