package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

var currentMatch *Match = nil

func matchRouter() http.Handler {
	r := chi.NewRouter()

	// Fetch the current match
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if currentMatch == nil {
			WriteErr(w, 400, "NoMatchGoingOn")
			return
		}

		json, _ := json.Marshal(currentMatch)
		w.WriteHeader(200)
		w.Write(json)
	})

	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != AUTH_CODE {
			WriteErr(w, 401, "Unauthorized")
			return
		}
		
		b, _ := ioutil.ReadAll(r.Body)
		
		body := NewMatchRequest{}

		err := json.Unmarshal(b, &body)
		if err != nil || body.Player1 == "" || body.Player2 == "" {
			WriteErr(w, 400, "BodyIsBad")
			return
		}

		if currentMatch != nil {
			WriteErr(w, 400, "MatchAlreadyStarted")
			return
		}

		player1, err1 := FetchPlayer(body.Player1, false)
		player2, err2 := FetchPlayer(body.Player2, false)

		if err1 != nil || err2 != nil {
			WriteErr(w, 400, "PlayerNotFound")
			return
		}

		currentMatch = &Match{
			P1:  player1,
			P2:  player2,
			P1P: 0,
			P2P: 0,
		}

		matchJSON, _ := json.Marshal(currentMatch)
		w.WriteHeader(200)
		w.Write(matchJSON)

		wsBroadcast <- Event{
			Event: "newMatch",
			Data:  matchJSON,
		}
	})

	r.Post("/end", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != AUTH_CODE {
			WriteErr(w, 401, "Unauthorized")
			return
		}
		
		if currentMatch == nil {
			WriteErr(w, 400, "NoMatchGoingOn")
			return
		}

		currentMatch = nil
		Respond(w, 200, `{"status":"success"}`)

		wsBroadcast <- Event{
			Event: "endMatch",
			Data:  nil,
		}
	})

	r.Get("/stake", func(w http.ResponseWriter, r *http.Request) {
		if currentMatch == nil {
			WriteErr(w, 400, "NoMatchGoingOn")
			return
		}

		stake := getStake(currentMatch.P1.ELO, currentMatch.P2.ELO)
		stakeResp := StakeResp{
			Stake: stake,
		}
		jsonEnc, _ := json.Marshal(stakeResp)
		w.WriteHeader(200)
		w.Write(jsonEnc)
	})

	r.Post("/player1", givePoint("1"))
	r.Post("/player2", givePoint("2"))
	r.Post("/draw", givePoint("3"))

	return r
}

func addELO(winner, looser Player, draw bool) (Player, Player) {
	probWinner := 1.0 / (1.0 + math.Pow(10, float64(looser.ELO-winner.ELO) / 400))

	actualProb := 1.0

	if draw {
		actualProb = 0.5
	}

	eloStake := int(math.Round(ELO_K*(actualProb-probWinner)))

	winner.ELO += eloStake
	looser.ELO -= eloStake

	return winner, looser
}

func getStake(elo1, elo2 int) int {
	prob := 1.0 / (1.0 + math.Pow(10, float64(elo1-elo2) / 400))
	if prob < 0.5 {
		prob = 1 - prob
	}
	return int(math.Abs(math.Round(ELO_K*(1-prob))))
}

func givePoint(player string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentMatch == nil {
			WriteErr(w, 400, "NoMatchGoingOn")
			return
		}
		winner := &Player{}
		looser := &Player{}
		switch player {
		case "1":
			winner = &currentMatch.P1
			looser = &currentMatch.P2
			currentMatch.P1P++
		case "2":
			winner = &currentMatch.P2
			looser = &currentMatch.P1
			currentMatch.P2P++
		case "3":
			if os.Getenv("DRAW_MODE") == "GIVE" {
				currentMatch.P1P++
				currentMatch.P2P++
			}
			winner = &currentMatch.P1
			looser = &currentMatch.P2
		default:
			WriteErr(w, 404, "NotFound")
		}
		
		win, loose := addELO(*winner, *looser, player == "3")
		*winner = win
		*looser = loose
		
		currentMatch.P1.UpdateSQL()
		currentMatch.P2.UpdateSQL()

		jsonEnc, _ := json.Marshal(currentMatch)
		w.WriteHeader(200)
		w.Write(jsonEnc)

		wsBroadcast <- Event{
			Event: "newPoint",
			Data:  jsonEnc,
		}

		return
	}
}