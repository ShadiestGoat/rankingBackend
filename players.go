package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
)


func playerRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		sql, _ := DB.Query(`SELECT id, name, elo FROM players ORDER BY elo DESC`)
		players := []Player{}
		for sql.Next() {
			id, name, elo := "", "", 0
			sql.Scan(&id, &name, &elo)
			players = append(players, Player{
				PlayerBase: PlayerBase{
					ELO: elo,
					ID:  id,
				},
				Name:       name,
			})
		}
		jsonEnc, _ := json.Marshal(players)
		w.WriteHeader(200)
		w.Write(jsonEnc)
	})

	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != AUTH_CODE {
			WriteErr(w, 401, "Unauthorized")
			return
		}

		b, _ := ioutil.ReadAll(r.Body)
		
		body := NewPlayerRequest{}

		err := json.Unmarshal(b, &body)
		if err != nil {
			WriteErr(w, 400, "BodyIsBad")
			return
		}

		if 2 >= len(body.Name) || len(body.Name) > 24 {
			WriteErr(w, 400, "BodyIsBad")
			return
		}

		if 100 >= body.ELO || body.ELO > 5000 {
			WriteErr(w, 400, "BodyIsBad")
			return
		}

		id := SnowNode.Generate().String()

		psqlInsert := fmt.Sprintf(`INSERT INTO players (elo, id, name) VALUES (%v, %v, '%v')`, body.ELO, id, body.Name)
		DB.Exec(psqlInsert)

		player := Player{
			PlayerBase: PlayerBase{
				ELO: body.ELO,
				ID:  id,
			},
			Name:       body.Name,
		}

		jsonEnc, _ := json.Marshal(player)

		w.WriteHeader(200)
		w.Write(jsonEnc)
	})

	r.Get("/profile/{PlayerID}", func(w http.ResponseWriter, r *http.Request) {
		playerID := chi.URLParam(r, "PlayerID")
		player, err := FetchPlayer(playerID, true)
		
		if err != nil {
			WriteErr(w, 400, "PlayerNotFound")
			return
		}

		jsonEnc, _ := json.Marshal(player)
		
		w.WriteHeader(200)
		w.Write(jsonEnc)
	})

	return r
}