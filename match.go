package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"

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

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
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
			initFavor := 1
	
			if player2.ELO > player1.ELO {
				initFavor = 2
			}
	
			if err1 != nil || err2 != nil {
				WriteErr(w, 400, "PlayerNotFound")
				return
			}
	
			currentMatch = &Match{
				P1:  player1,
				P2:  player2,
				P1P: 0,
				P2P: 0,
				InitialFavor: initFavor,
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
			if currentMatch == nil {
				WriteErr(w, 400, "NoMatchGoingOn")
				return
			}
	
			var winner *Player
			var looser *Player
			var winnerID int
	
			var stake int
	
			if currentMatch.P1P > currentMatch.P2P {
				winner = &currentMatch.P1
				looser = &currentMatch.P2
				winnerID = 1
			} else {
				winner = &currentMatch.P2
				looser = &currentMatch.P1
				winnerID = 2
			}
	
			if winnerID == currentMatch.InitialFavor {
				stake = int(math.Round(float64(ELO_K)*1.75))
			} else {
				stake = int(math.Round(float64(ELO_K)*0.25))
			}
	
			winner.ELO += stake
			looser.ELO -= stake
	
			winner.UpdateSQL()
			looser.UpdateSQL()
	
			currentMatch = nil
			Respond(w, 200, `{"status":"success"}`)
	
			wsBroadcast <- Event{
				Event: "endMatch",
				Data:  nil,
			}
		})

		r.Post("/player1", givePoint("1"))
		r.Post("/player2", givePoint("2"))
		r.Post("/draw", givePoint("3"))
	})

	return r
}

type AddEloDrawOptions struct {
	IsDraw bool
	IsIntentionalDraw bool
}

func addELO(winner, looser *Player, draw AddEloDrawOptions) {
	probWinner := 1.0 / (1.0 + math.Pow(10, float64(looser.ELO-winner.ELO) / 400))

	actualProb := 1.0

	if draw.IsDraw {
		actualProb = 0.5
		if draw.IsIntentionalDraw {
			actualProb = 0.75
		}
	}

	eloStake := int(math.Round(ELO_K*(actualProb-probWinner)))

	winner.ELO += eloStake
	looser.ELO -= eloStake
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
		drawOpts := AddEloDrawOptions{}

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
			drawOpts.IsDraw = true
			winner = &currentMatch.P1
			looser = &currentMatch.P2
			
			if GIVE_ON_DRAW {
				currentMatch.P1P++
				currentMatch.P2P++
				drawOpts.IsIntentionalDraw = currentMatch.P1P != currentMatch.P2P
				if drawOpts.IsIntentionalDraw {
					if currentMatch.P1P < currentMatch.P2P {
						winner = &currentMatch.P2
						looser = &currentMatch.P1
					}
				}
			}
		default:
			WriteErr(w, http.StatusNotFound, "NotFound")
		}
		
		addELO(winner, looser, drawOpts)
		
		currentMatch.P1.UpdateSQL()
		currentMatch.P2.UpdateSQL()

		jsonEnc, _ := json.Marshal(currentMatch)
		w.WriteHeader(200)
		w.Write(jsonEnc)

		wsBroadcast <- Event{
			Event: "newPoint",
			Data:  jsonEnc,
		}
	}
}