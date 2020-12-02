package middleware_test

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	. "go.krishnaiyer.dev/go-vanity-docker/pkg/middleware"

	"github.com/gorilla/mux"
)

func TestBasicAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	address := "localhost:8080"
	r := mux.NewRouter()
	r.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})
	s := &http.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}
	auth := NewBasicAuth("test", "secret")
	auth.Add("test1", "secret1")
	r.Use(auth.Middleware("test"))

	go func() {
		log.Printf("Starting web server at %s\n", address)
		select {
		case <-ctx.Done():
			s.Close()
		default:
			log.Fatal(s.ListenAndServe().Error())
		}
	}()

	// Make requests with/without auth
	// Check for empty password with invalid username

}
