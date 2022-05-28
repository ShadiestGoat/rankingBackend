package main

import "time"

type PlayerBase struct {
	ELO int `json:"elo"`
	ID string `json:"id"`
}

type Player struct {
	PlayerBase
	Name string `json:"name"`
	History []History `json:"history,omitempty"`
}

type History struct {
	PlayerBase
	Date time.Time
}

type Match struct {
	P1 Player `json:"player1"`
	P2 Player `json:"player2"`
	P1P int `json:"player1Points"`
	P2P int `json:"player2Points"`
}

type NewMatchRequest struct {
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
}

type NewPlayerRequest struct {
	Name string `json:"name"`
	ELO int `json:"elo"`
}

type StakeResp struct {
	Stake int `json:"stake"`
}