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
	"fmt"
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
	Redirects    string `name:"redirects" short:"r" description:"remote URL or local path to vanity redirects as yml"`
	HTTPAddress  string `name:"http-address" short:"a" description:"host:port for the http server"`
	Debug        bool   `name:"debug" short:"d" description:"print detailed logs for errors"`
	NoOfSubPaths int    `name:"no-of-subpaths" short:"n" description:"number of subpaths"`
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

			h, err := handler.Init(ctx, config.Redirects)
			if err != nil {
				log.Fatal(err.Error())
			}
			var address string
			if address = addressRegex.FindString(config.HTTPAddress); address == "" {
				log.Println(fmt.Sprintf("Invalid or empty server address %s using 0.0.0.0:8080", config.HTTPAddress))
				address = "0.0.0.0:8080"
			}

			r := mux.NewRouter()
			r.HandleFunc("/", h.HandleIndex)
			addPath(r, h.HandleImport, config.NoOfSubPaths)
			r.Methods("GET")
			s := &http.Server{
				Handler:      r,
				Addr:         address,
				WriteTimeout: 5 * time.Second,
				ReadTimeout:  5 * time.Second,
				IdleTimeout:  5 * time.Second,
			}

			log.Println(fmt.Sprintf("Serving HTTP requests on %s ", address))
			select {
			case <-ctx.Done():
				s.Close()
			default:
				log.Fatal(s.ListenAndServe().Error())
			}
		},
	}
)

func addPath(r *mux.Router, f func(http.ResponseWriter, *http.Request), n int) {
	path := "/{project}"
	r.HandleFunc(path, f)
	for i := 0; i < n; i++ {
		path = fmt.Sprintf("%s/{path%d}", path, i)
		r.HandleFunc(path, f)
		log.Println(fmt.Sprintf("Adding Path %s", path))
	}
}

// Execute ...
func Execute() {
	if err := Root.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func init() {
	manager = New("config", "go-vanity")
	manager.InitFlags(*config)
	Root.PersistentFlags().AddFlagSet(manager.Flags())
	Root.AddCommand(manager.VersionCommand(Root))
	Root.AddCommand(manager.ConfigCommand(Root))
}
