package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	serviceconfig "github.com/momentohq/acorn-smash-demo/internal/config"

	"github.com/momentohq/client-sdk-go/momento"
	"github.com/momentohq/client-sdk-go/responses"
	"github.com/momentohq/client-sdk-go/utils"
)

type GameController struct {
	MomentoClient momento.CacheClient
}

type buttonHitRequest struct {
	User string `json:"user"`
}

type scoreBoardEntry struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
type scoreBoardResponse struct {
	Elements []scoreBoardEntry `json:"elements"`
}

func (c *GameController) RegisterHit(w http.ResponseWriter, r *http.Request) {
	var request buttonHitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeFatalError(w, "fatal error occurred decoding msg payload", err)
	}
	_, err := c.MomentoClient.SortedSetIncrementScore(r.Context(), &momento.SortedSetIncrementScoreRequest{
		CacheName: serviceconfig.CacheName,
		SetName:   "score-board",
		Value:     momento.String(request.User),
		Amount:    1,
		Ttl: &utils.CollectionTtl{
			Ttl:        24 * time.Hour,
			RefreshTtl: true,
		},
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred incrementing user score", err)
	}
}

func (c *GameController) GetTopScorers(w http.ResponseWriter, r *http.Request) {
	var count uint32 = 10
	resp, err := c.MomentoClient.SortedSetFetchByScore(r.Context(), &momento.SortedSetFetchByScoreRequest{
		CacheName: serviceconfig.CacheName,
		SetName:   "score-board",
		Order:     momento.DESCENDING,
		Count:     &count,
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred getting top scores", err)
	}
	var scoreBoardEntries []scoreBoardEntry
	switch r := resp.(type) {
	case *responses.SortedSetFetchHit:
		for _, e := range r.ValueStringElements() {
			scoreBoardEntries = append(scoreBoardEntries, scoreBoardEntry{
				Name:  e.Value,
				Value: e.Score,
			})
		}
	}

	if err := json.NewEncoder(w).Encode(&scoreBoardResponse{Elements: scoreBoardEntries}); err != nil {
		writeFatalError(w, "fatal error getting score", err)
		return
	}
}
