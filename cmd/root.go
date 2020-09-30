// Copyright © 2020 Krishna Iyer Easwaran
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.krishnaiyer.dev/go-vanity-docker/pkg/handler"
)

// Config represents the configuration
type Config struct {
	VanityConfig string `name:"vanity-config" short:"c" description:"remote URL or local path to vanity configuration as yml"`
	HTTPAddress  string `name:"http-address" short:"a" description:"host:port for the http server"`
	Debug        bool   `name:"debug" short:"d" description:"print detailed logs for errors"`
}

var (
	flags = pflag.NewFlagSet("go-vanity", pflag.ExitOnError)

	config = new(Config)

	manager *Manager

	addressRegex = regexp.MustCompile(`^([a-z-.0-9]+)(:[0-9]+)?$`)

	// Root is the root of the commands.
	Root = &cobra.Command{
		Use:           "go-vanity",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "go-vanity handles go vanity redirect requests",
		Long:          `go-vanity handles go vanity redirect requests and is available for deployment via docker. More documentation at https://go.krishnaiyer.dev/go-vanity-docker`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := manager.Unmarshal(config)
			if err != nil {
				panic(err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			baseCtx := context.Background()
			ctx, cancel := context.WithCancel(baseCtx)
			defer cancel()

			h, err := handler.Init(ctx, config.VanityConfig)
			if err != nil {
				log.Fatal(err)
			}
			var address string
			if address = addressRegex.FindString(config.HTTPAddress); address == "" {
				log.Printf("Invalid or empty server address %s using 0.0.0.0:8080", config.HTTPAddress)
				address = "0.0.0.0:8080"
			}

			r := mux.NewRouter()
			r.HandleFunc("/", h.HandleIndex)
			r.HandleFunc("/{project}", h.HandleImport)
			r.HandleFunc("/{project}/{path}", h.HandleImport)
			r.Methods("GET")
			s := &http.Server{
				Handler:      r,
				Addr:         address,
				WriteTimeout: 5 * time.Second,
				ReadTimeout:  5 * time.Second,
				IdleTimeout:  5 * time.Second,
			}

			log.Printf("Serving HTTP requests on %s ", address)
			select {
			case <-ctx.Done():
				s.Close()
			default:
				log.Fatal(s.ListenAndServe())
			}
		},
	}
)

// Execute ...
func Execute() {
	if err := Root.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func init() {
	manager = New("config")
	manager.InitFlags(*config)
	Root.PersistentFlags().AddFlagSet(manager.Flags())
	Root.AddCommand(Version(Root))
}
