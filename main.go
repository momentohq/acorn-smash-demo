package main

import (
	"net/http"
	"time"

	"github.com/eaddingtonwhite/momento-game-demo/internal/controllers"

	"github.com/momentohq/client-sdk-go/auth"
	"github.com/momentohq/client-sdk-go/config"
	"github.com/momentohq/client-sdk-go/momento"
)

func main() {
	credProvider, err := auth.NewEnvMomentoTokenProvider("MOMENTO_AUTH_TOKEN")
	if err != nil {
		panic(err)
	}
	client, err := momento.NewSimpleCacheClient(&momento.SimpleCacheClientProps{
		Configuration:      config.LatestLaptopConfig(),
		CredentialProvider: credProvider,
		DefaultTTL:         60 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	chatController := &controllers.ChatController{
		MomentoClient: client,
	}

	gameController := &controllers.GameController{
		MomentoClient: client,
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
