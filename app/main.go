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

func newRouter() *httprouter.Router {
	ytApiKey := os.Getenv("YOUTUBE_API_KEY")
	ytChannelID := os.Getenv("YOUTUBE_CHANNEL_ID")
	if ytApiKey == "" {
		log.Fatalf("No youtube key")
	}
	if ytChannelID == "" {
		log.Fatalf("No cannel id")
	}

	mux := httprouter.New()
	mux.GET("/youtube/channel/stats", getChannelStats(ytApiKey, ytChannelID))

	return mux
}

func main() {
	srv := &http.Server{
		Handler: newRouter(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint
		log.Println("service interrupt received")

		ctx, cancle := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancle()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}

		log.Println("shutdown complete")

		close(idleConnsClosed)

	}()

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("fatal http server failed to start: %v", err)
		}
	}

	<-idleConnsClosed
	log.Println("Service Stop")
}
