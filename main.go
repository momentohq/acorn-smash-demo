package main

import (
	"github.com/momentohq/acorn-smash-demo/internal/controllers"
	"net/http"
	"time"

	"github.com/momentohq/client-sdk-go/auth"
	"github.com/momentohq/client-sdk-go/config"
	"github.com/momentohq/client-sdk-go/momento"
)

func main() {
	credProvider, err := auth.NewEnvMomentoTokenProvider("MOMENTO_AUTH_TOKEN")
	if err != nil {
		panic(err)
	}
	topicClient, err := momento.NewTopicClient(
		config.LaptopLatest(),
		credProvider,
	)
	if err != nil {
		panic(err)
	}
	chatController := &controllers.ChatController{
		MomentoTopicClient: topicClient,
	}

	cacheClient, err := momento.NewCacheClient(
		config.LaptopLatest(),
		credProvider,
		60*time.Second,
	)
	gameController := &controllers.GameController{
		MomentoClient: cacheClient,
	}

	http.HandleFunc("/connect", chatController.Connect)
	http.HandleFunc("/send-message", chatController.SendMessage)

	http.HandleFunc("/register-hit", gameController.RegisterHit)
	http.HandleFunc("/top-scorers", gameController.GetTopScorers)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
