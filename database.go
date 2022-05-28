package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dextication/snowflake"

	"database/sql"

	_ "github.com/lib/pq"
)

var (
	SnowNode *snowflake.Node
	DB *sql.DB
	BASE_ID_TIME     = time.Date(2021, time.June, 11, 0, 0, 0, 0, time.UTC)
	BASE_ID_STAMP    = BASE_ID_TIME.UnixMilli()
)

const PSQL_DATE_LAYOUT = "2006-01-02"

const createHistoryTable = `CREATE TABLE IF NOT EXISTS history (
elo NUMERIC(5,0),
id VARCHAR(18),
date DATE,

CONSTRAINT is_payer
	FOREIGN KEY(id)
		REFERENCES players(id)
);`

const createPlayerTable = `CREATE TABLE IF NOT EXISTS players (
elo NUMERIC(5,0),
name VARCHAR(24),
id VARCHAR(18),

PRIMARY KEY(id)
);`

func InitDB() {
	node, err := snowflake.NewNode(0, BASE_ID_TIME, 41, 11, 11)
	PanicIfErr(err)
	SnowNode = node

	// Connect to the db
	psqlconn := fmt.Sprintf("host=%s port=%v user=%s password=%v dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME)

	db, err := sql.Open("postgres", psqlconn)
	PanicIfErr(err)

	err = db.Ping()
	PanicIfErr(err)

	_, err = db.Exec(createPlayerTable)
	PanicIfErr(err)

	_, err = db.Exec(createHistoryTable)
	PanicIfErr(err)

	DB = db
}

func FetchPlayer(id string, fetchHistory bool) (Player, error) {
	player := Player{
		PlayerBase: PlayerBase{
			ID:  id,
		},
	}

	qBase := fmt.Sprintf("SELECT name, elo FROM players WHERE id = '%v'", id)
	baseResp := DB.QueryRow(qBase)
	PanicIfErr(baseResp.Err())
	err := baseResp.Scan(&player.Name, &player.ELO)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Player{}, ErrPlayerNotFound
		}
		return Player{}, nil
	}
	
	qHist := fmt.Sprintf("SELECT elo, date FROM history WHERE id = '%v'", id)
	histResp, err := DB.Query(qHist)
	PanicIfErr(err)
	PanicIfErr(baseResp.Err())

	for histResp.Next() {
		elo := 0
		date := ""

		histResp.Scan(&elo, &date)
		
		trueDate, _ := time.Parse(PSQL_DATE_LAYOUT, date)

		player.History = append(player.History, History{
			PlayerBase: PlayerBase{
				ELO: elo,
				ID:  id,
			},
			Date:       trueDate,
		})
	}

	return player, nil
}

func (p Player) UpdateSQL() {
	DB.Exec(fmt.Sprintf(`UPDATE players SET elo = '%v' WHERE id = '%v'`, p.ELO, p.ID))
}
