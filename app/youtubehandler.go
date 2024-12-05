package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeStats struct {
	Subscribers    int    `json:"subscribers"`
	ChannelName    string `json:"channelName"`
	MinutesWatched int    `json:"minutesWatched,omitempty"`
	Views          int    `json:"views"`
}

var youtubeService *youtube.Service
var once sync.Once

func getYouTubeService(ctx context.Context, apiKey string) (*youtube.Service, error) {
	var err error
	once.Do(func() {
		youtubeService, err = youtube.NewService(ctx, option.WithAPIKey(apiKey))
	})
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %w", err)
	}
	return youtubeService, nil
}

func fetchChannelStats(_ context.Context, yts *youtube.Service, channelID string) (*youtube.ChannelListResponse, error) {
	call := yts.Channels.List([]string{"snippet", "contentDetails", "statistics"})
	response, err := call.Id(channelID).Do()
	if err != nil {
		return nil, fmt.Errorf("error making YouTube API call: %w", err)
	}
	return response, nil
}

func handleChannelStats(w http.ResponseWriter, _ *http.Request, _ httprouter.Params, k, channelID string) {
	ctx := context.Background()
	yts, err := getYouTubeService(ctx, k)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response, err := fetchChannelStats(ctx, yts, channelID)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var yt YoutubeStats
	if len(response.Items) > 0 {
		val := response.Items[0]
		yt = YoutubeStats{
			ChannelName: val.Snippet.Title,
			Subscribers: int(val.Statistics.SubscriberCount),
			Views:       int(val.Statistics.ViewCount),
		}
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(yt); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getChannelStats(k, channelID string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		handleChannelStats(w, r, ps, k, channelID)
	}
}
