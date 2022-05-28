package main

import "net/http"

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteErr(w http.ResponseWriter, status int, msg string) {
	Respond(w, status, `{"err":"` + msg + `"}`)
}

func Respond(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}