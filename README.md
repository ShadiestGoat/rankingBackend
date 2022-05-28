# Ranker Backend

## What is this?

This is a backend server for competative 1v1 matches. It stores [ELO](https://en.wikipedia.org/wiki/Elo_rating_system) of each player, as a means of ranking each player's skill. It also stores a backup of each player's elo with configurable frequency, so that one could create a history of a player's ELO. 

## Setup

This project uses Go and PostgresSQL. So you need a PostgresSQL database running, and you need Go of version 1.18+. After that setup, create a `.env` file in a seperate folder, and install this project `go install github.com/ShadiestGoat/rankingBackend`

## Configuration

This project is configured through env variables. It supports a `.env` file. There is a `template.env` file with senseble defaults as well

| ENV VAR | Description | Default |
|:-:|:-:|:-:|
| DB_HOST | The host of the database | :x: |
| DB_PORT | The port of the database | 5432 |
| DB_USER | The username of the user for the database | :x: |
| DB_PASS | The password of the user for the database | :x: |
| DB_NAME | The database name | :x: |
| AUTH_CODE | An authentication code used by the front end, to identify admins | :x: |
| DRAW_MODE | What is done if there is a draw. The 2 options are GIVE which would give points to both players, vs any other string, which would just not give points | "GIVE" |
| BACKUP_DAYS | A string representation of when to do player ELO backups. The default is daily. The syntax is a string of < 8 in length, each character representing each day. `*` would mean to backup, and ` ` (space) means don't. The first character represents **Sunday**. If a character is not present, it's default is false. (so `**` means to only backup on sunday and monday) | `*******` |
| ELO_K | A number. Its the multiplicant used in the ELO formula | 50 |
| PORT | The port that this project is hosted on | 3000 |

## Running

Make sure that the `.env` file is in the working directory. If you have installed the project as said in [the setup guide](#setup), then simply run `rankingBackend`.
This not only runs the backend, but also hosts the frontend. This frontend is located in the `frontend` folder. It specifically hosts statically built html. Here is [the frontend portion of this project](https://github.com/ShadiestGoat/rankingFrontend). See it's guide for building, etc.
