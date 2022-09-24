package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron"
)

func main() {
	InitConfig()
	InitDB()

	r := chi.NewRouter()

	r.Mount("/api", Router())

	wd, _ := os.Getwd()
	dir := http.Dir(filepath.Join(wd + "/frontend"))

	FileServer(r, "/", dir)

	s := gocron.NewScheduler(time.UTC).Week()

	for i, char := range BACKUP_DAYS {
		if i > 6 {break}
		
		switch string(char) {
		case "*":
			s = s.Weekday(time.Weekday(i))
		case " ":
		default:
			panic(fmt.Sprintf("BACKUP_DAYS isn't correct! Charecter #%v is '%v' which isn't ' ' or '*'!", i, string(char)))
		}
	}

	s.At("23:59").Do(func () {
		playersRet, _ := DB.Query(`SELECT elo, id FROM players`)
		if err := playersRet.Err(); err != nil {
			panic(err)
		}
		now := time.Now().Format(PSQL_DATE_LAYOUT)

		for playersRet.Next() {
			elo := 0
			id := ""

			playersRet.Scan(&elo, &id)
			DB.Exec(fmt.Sprintf(`INSERT INTO history (elo, id, date) VALUES (%v, '%v', '%v')`, elo, id, now))
		}
	})

	go channelReceiver()
	
	fmt.Println("Server Ready at http://localhost:" + PORT + "!")
	http.ListenAndServe(":" + PORT, r)
}
