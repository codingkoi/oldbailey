package view

import (
	"../record"
	"../search"
	auth "github.com/abbot/go-http-auth"
	"github.com/hoisie/mustache"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func SearchHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	switch r.Method {
	case "GET":
		params := r.URL.Query()
		page, err := strconv.Atoi(params.Get("page"))
		if err != nil {
			page = 1
		}
		searchText := params.Get("text")

		responseBody := mustache.RenderFile("view/search.html",
			search.Search(searchText, page))
		w.Write([]byte(responseBody))
	default:
		methodNotAllowed(w, &r.Request)
	}
}

func CaseHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	var err error
	path := r.URL.Path
	pattern, err := regexp.Compile("/case/([^/]*)")
	if err != nil {
		panic(err)
	}
	matches := pattern.FindStringSubmatch(path)
	if len(matches) != 2 {
		resourceNotFound(w, &r.Request)
		return
	}
	id := matches[1]

	switch r.Method {
	case "GET":
		baileyCase := record.FetchRecord(id)
		if baileyCase == nil {
			resourceNotFound(w, &r.Request)
			return
		}
		responseBody := mustache.RenderFile("view/case.html", baileyCase)
		w.Write([]byte(responseBody))

	case "POST":
		baileyCase := record.FetchRecord(id)
		if baileyCase == nil {
			resourceNotFound(w, &r.Request)
			return
		}
		baileyCase.OfInterest, err = strconv.ParseBool(r.FormValue("of-interest"))
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}
		note := &baileyCase.Note
		note.Notes = r.FormValue("notes")
		note.Clothing, err = strconv.ParseBool(r.FormValue("clothing"))
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}
		note.RawTextiles, err = strconv.ParseBool(r.FormValue("raw-textiles"))
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}
		note.OtherTextiles, err = strconv.ParseBool(r.FormValue("other-textiles"))
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}
		note.Other, err = strconv.ParseBool(r.FormValue("other"))
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}
		baileyCase.Save()
		successNoContent(w, &r.Request)

	default:
		methodNotAllowed(w, &r.Request)
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		url := r.URL
		path := url.Path[1:]
		http.ServeFile(w, r, path)
	default:
		methodNotAllowed(w, r)
	}
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	body := "Method Not Allowed: " + r.Method
	w.Write([]byte(body))
}

func badRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	body := "Bad Request"
	w.Write([]byte(body))
}

func resourceNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	body := "Resource Not Found"
	w.Write([]byte(body))
}

func successNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
