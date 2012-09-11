package view

import (
	"../record"
	"../search"
	"encoding/json"
	auth "github.com/abbot/go-http-auth"
	"github.com/hoisie/mustache"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type JsonUpdate struct {
	OfInterest          bool
	NotOfInterest       bool
	Notes               string
	Clothing            bool
	ClothingCount       int
	RawTextiles         bool
	RawTextilesCount    int
	OtherTextiles       bool
	Accessories         bool
	AccessoriesCount    int
	HouseholdLinen      bool
	HouseholdLinenCount int
	Other               bool
	OtherCount          int
	OtherNotSpecified   bool
}

func SearchHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	switch r.Method {
	case "GET":
		params := r.URL.Query()
		page, err := strconv.Atoi(params.Get("page"))
		if err != nil {
			page = 1
		}
		searchText := params.Get("text")

		responseContent := search.Search(searchText, page)
		var responseBody []byte
		if strings.Contains(r.Header["Accept"][0], "application/json") {
			responseBody, err = json.Marshal(responseContent)
			if err != nil {
				internalServerError(w, &r.Request)
				return
			}
		} else {
			responseBody = []byte(mustache.RenderFile("view/search.html", responseContent))
		}
		w.Write(responseBody)
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
		var jsonUpdate JsonUpdate
		err = json.Unmarshal([]byte(r.FormValue("json")), &jsonUpdate)
		if err != nil {
			log.Println(err)
			badRequest(w, &r.Request)
			return
		}

		baileyCase.OfInterest = jsonUpdate.OfInterest
		baileyCase.NotOfInterest = jsonUpdate.NotOfInterest
		baileyCase.Clothing = jsonUpdate.Clothing
		baileyCase.ClothingCount = jsonUpdate.ClothingCount
		baileyCase.RawTextiles = jsonUpdate.RawTextiles
		baileyCase.RawTextilesCount = jsonUpdate.RawTextilesCount
		baileyCase.OtherTextiles = jsonUpdate.OtherTextiles
		baileyCase.Accessories = jsonUpdate.Accessories
		baileyCase.AccessoriesCount = jsonUpdate.AccessoriesCount
		baileyCase.HouseholdLinen = jsonUpdate.HouseholdLinen
		baileyCase.HouseholdLinenCount = jsonUpdate.HouseholdLinenCount
		baileyCase.Other = jsonUpdate.Other
		baileyCase.OtherCount = jsonUpdate.OtherCount
		baileyCase.OtherNotSpecified = jsonUpdate.OtherNotSpecified

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

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	body := "Internal Server Error"
	w.Write([]byte(body))
}
