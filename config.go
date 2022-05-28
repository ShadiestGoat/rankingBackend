package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type confItem struct {
	Res         *string
	Default     string
	PanicIfNone bool
}


var (
	DB_HOST = ""
	DB_PORT = ""
	DB_USER = ""
	DB_PASS = ""
	DB_NAME = ""
	AUTH_CODE = ""
	BACKUP_DAYS = ""
	ELO_K = 0.0
	PORT = ""
)

func InitConfig() {
	godotenv.Load(".env")

	ELO_STR := ""

	var confMap = map[string]confItem{
		"AUTH_CODE": {
			PanicIfNone: true,
			Res:         &AUTH_CODE,
		},
		"BACKUP_DAYS": {
			Default: "*******",
			Res: &BACKUP_DAYS,
		},
		"PORT": {
			Res: &PORT,
			Default: "3000",
		},
		"DB_HOST": {
			Res: &DB_HOST,
			PanicIfNone: true,
		},
		"DB_PORT": {
			Res: &DB_PORT,
			Default: "5432",
		},
		"DB_USER": {
			Res: &DB_USER,
			PanicIfNone: true,
		},
		"DB_PASS": {
			Res: &DB_PASS,
			PanicIfNone: true,
		},
		"DB_NAME": {
			Res: &DB_NAME,
			PanicIfNone: true,
		},
		"ELO_K": {
			Res: &ELO_STR,
			Default:     "50",
		},
	}

	for name, opt := range confMap {
		item := os.Getenv(name)
		
		if len(item) == 0 {
			if opt.PanicIfNone {
				panic(fmt.Sprintf("'%v' is a needed variable, but is not present! Please read the README.md file for more info.", name))
			}
			item = opt.Default
		}

		*opt.Res = item
	}

	eloK, err := strconv.ParseFloat(ELO_STR, 64)
	PanicIfErr(err)
	ELO_K = eloK
}
