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

package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestHandler(t *testing.T) {
	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	a := assertions.New(t)
	h, err := InitHandler(ctx, "./test.yml")
	a.So(err, should.BeNil)

	address := "127.0.0.1:8000"

	r := mux.NewRouter()
	r.HandleFunc("/", h.HandleIndex)
	r.HandleFunc("/{project}", h.HandleImport)
	r.Methods("GET")
	s := &http.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}
	go func() {
		select {
		case <-ctx.Done():
			s.Close()
		default:
			log.Fatal(s.ListenAndServe())
		}
	}()

	for _, tc := range []struct {
		Name                string
		URL                 string
		ExpectedStatusCode  int
		ExpectedContentType []string
		ExpectedBody        string
	}{
		{
			Name:                "Index",
			URL:                 fmt.Sprintf("http://%s/", address),
			ExpectedContentType: []string{"text/html; charset=utf-8"},
			ExpectedStatusCode:  http.StatusOK,
			ExpectedBody: `<!DOCTYPE html>
<html>
<h1>Welcome to localhost</h1>
<ul>
<li><a href="https://pkg.go.dev/localhost/mycoolproject">localhost/mycoolproject</a></li><li><a href="https://pkg.go.dev/localhost/myothercoolproject">localhost/myothercoolproject</a></li>
</ul>
</html>
`,
		},
		{
			Name:                "Project1",
			URL:                 fmt.Sprintf("http://%s/mycoolproject", address),
			ExpectedStatusCode:  http.StatusOK,
			ExpectedContentType: []string{"text/html; charset=utf-8"},
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="localhost/mycoolproject git https://github.com/user/mycoolproject">
<meta name="go-source" content="localhost/mycoolproject https://github.com/user/mycoolproject https://github.com/user/mycoolproject/tree/master{/dir} https://github.com/user/mycoolproject/blob/master{/dir}/{file}#L{line}">
</head>
<body>
Nothing to see here folks!
</body>
</html>`,
		},
		{
			Name:                "Project2",
			URL:                 fmt.Sprintf("http://%s/myothercoolproject", address),
			ExpectedStatusCode:  http.StatusOK,
			ExpectedContentType: []string{"text/html; charset=utf-8"},
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="localhost/myothercoolproject git https://github.com/user/myothercoolproject">
<meta name="go-source" content="localhost/myothercoolproject https://github.com/user/myothercoolproject https://github.com/user/myothercoolproject/tree/master{/dir} https://github.com/user/myothercoolproject/blob/master{/dir}/{file}#L{line}">
</head>
<body>
Nothing to see here folks!
</body>
</html>`,
		},
		{
			Name:                "InvalidPath",
			URL:                 fmt.Sprintf("http://%s/myunknownproject", address),
			ExpectedContentType: []string{"text/plain; charset=utf-8"},
			ExpectedStatusCode:  http.StatusNotFound,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := http.Get(tc.URL)
			a.So(err, should.BeNil)
			a.So(tc.ExpectedStatusCode, should.Equal, resp.StatusCode)
			a.So(resp.Header["Content-Type"], should.Resemble, tc.ExpectedContentType)
			if resp.StatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(resp.Body)
				a.So(err, should.BeNil)
				a.So(string(body), should.Equal, tc.ExpectedBody)
			}
		})
	}
}
