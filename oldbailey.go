package main

import (
	"./record"
	"./view"
	"fmt"
	"net/http"
	"os"
	"log"
	auth "github.com/abbot/go-http-auth"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-r" {
		baileyCase := record.FetchRecord("t17000115-4")
		fmt.Println(string(baileyCase.RawText))
	}

	if len(os.Args) == 2 && os.Args[1] == "-s" {
		runServer()
		return
	}
}

func runServer() {
	secrets := auth.HtpasswdFileProvider(".htpasswd")
	a := auth.BasicAuthenticator("oldbailey", secrets)

	http.HandleFunc("/static/", view.StaticHandler)
	http.HandleFunc("/case/", a(view.CaseHandler))
	http.HandleFunc("/search", a(view.SearchHandler))
	log.Println("Staring server on port 2258")
	log.Fatal(http.ListenAndServe(":2258", nil))
}
