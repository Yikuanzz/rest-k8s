package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewRouter() *httprouter.Router {
	mux := httprouter.New()
	ytApiKey := os.Getenv("YOUTUBE_API_KEY")
	if ytApiKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable not set")
	}

	ytchannelID := os.Getenv("YOUTUBE_CHANNEL_ID")
	if ytApiKey == "" {
		log.Fatal("YOUTUBE_CHANNEL_ID environment variable not set")
	}
	// ytApiKey := "AIzaSyCwNuoD46cE7KGakrTgjkr5CRU5Z7fuGgU"
	// ytchannelID := "UClSv7tWDA4wkCTLhZl1YBlw"
	mux.GET("/youtube/channel/stats", getChannelStats(ytApiKey, ytchannelID))
	return mux
}

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: NewRouter(),
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint

		log.Println("Received an interrupt signal, shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP server Shutdown error: %v", err)
		}

		log.Println("HTTP server shutdown complete")
		close(idleConnsClosed)
	}()

	log.Printf("Service Start on port %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("fatal http server error: %v", err)
		}
	}

	// Wait for all connections to be closed
	<-idleConnsClosed
	log.Println("Service Stop")
}
