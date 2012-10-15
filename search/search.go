package search

import (
	"../record"
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	"math"
	"net/http"
	"net/url"
)

type searchResponse struct {
	Total int
	Hits  []string
}

type Results struct {
	Page            int
	Start           int
	End             int
	Count           int
	ResultsPerPage  int
	Total           int
	SearchText      string
	SubCategory     string
	Records         []*record.Record
	PaginationLinks []PaginationLink
}

// Search using the Old Bailey API to get the results
func Search(text string, page int) (results Results) {
	searchUrl, err := url.Parse("http://www.oldbaileyonline.org/obapi/ob")
	if err != nil {
		panic(err)
	}
	params := searchUrl.Query()
	params.Add("term0", "offcat_theft")
	params.Add("term1", "fromdate_17000115")
	params.Add("term2", "todate_17991204")
	if text != "" {
		params.Add("term3", "trialtext_"+text)
		results.SearchText = text
	}
	count := 15
	start := (page - 1) * count
	params.Add("start", fmt.Sprintf("%d", start))
	params.Add("count", fmt.Sprintf("%d", count))

	searchUrl.RawQuery = params.Encode()

	resp, err := http.Get(searchUrl.String())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var jsonData searchResponse
	err = json.Unmarshal(content, &jsonData)
	if err != nil {
		panic(err)
	}

	results.Count = len(jsonData.Hits)
	results.Total = jsonData.Total
	results.Start = start
	results.End = start + results.Count
	results.Page = page
	results.ResultsPerPage = count

	for _, hit := range jsonData.Hits {
		results.Records = append(results.Records, record.FetchRecord(hit))
	}
	results.SetPaginationLinks("/search")

	return
}

// Get a list of results from the cache of cases OfInterest and in the
// sub category specified
func CacheSearch(subcat string, page int) (results Results) {
	count := 15
	start := (page - 1) * count
	cachedRes := record.FetchSavedCases(subcat, start, count)

	results.Count = cachedRes.Count
	results.Total = cachedRes.Total
	results.Start = start
	results.End = cachedRes.Count + start
	results.Page = page
	results.ResultsPerPage = count
	results.SubCategory = subcat

	for _, hit := range cachedRes.Hits {
		results.Records = append(results.Records, record.FetchRecord(hit))
	}
	results.SetPaginationLinks("/cache")

	return
}

type PaginationLink struct {
	Href        string
	Label       string
	CurrentPage bool
}

func (results *Results) SetPaginationLinks(basepath string) {
	links := make([]PaginationLink, 0)
	baseUrl, _ := url.Parse(basepath)
	params := baseUrl.Query()
	if results.SearchText != "" {
		params.Set("text", results.SearchText)
	}
	if results.SubCategory != "" {
		params.Set("subcat", results.SubCategory)
	}
	baseUrl.RawQuery = params.Encode()

	curPage := float64(results.Page)
	lastPage := math.Ceil(float64(results.Total/results.ResultsPerPage)) + 1
	min := math.Max(curPage-5, 1)
	spread := math.Abs(min - curPage + 4)
	max := math.Min(curPage+spread+4, lastPage)

	// First and Prev links
	if curPage != 1 {
		params := baseUrl.Query()
		params.Set("page", fmt.Sprintf("%v", 1))
		baseUrl.RawQuery = params.Encode()
		links = append(links, PaginationLink{
			Href:  baseUrl.String(),
			Label: "First",
		})

		params.Set("page", fmt.Sprintf("%v", curPage-1))
		baseUrl.RawQuery = params.Encode()
		links = append(links, PaginationLink{
			Href:  baseUrl.String(),
			Label: "Prev",
		})
	}

	for i := int(min); i <= int(max); i++ {
		page := fmt.Sprintf("%v", i)
		params := baseUrl.Query()
		params.Set("page", page)
		baseUrl.RawQuery = params.Encode()
		links = append(links, PaginationLink{
			Href:        baseUrl.String(),
			Label:       page,
			CurrentPage: int(curPage) == i,
		})
	}

	// Last and Next links
	if curPage != lastPage {
		params := baseUrl.Query()
		params.Set("page", fmt.Sprintf("%v", curPage+1))
		baseUrl.RawQuery = params.Encode()
		links = append(links, PaginationLink{
			Href:  baseUrl.String(),
			Label: "Next",
		})

		params.Set("page", fmt.Sprintf("%v", lastPage))
		baseUrl.RawQuery = params.Encode()
		links = append(links, PaginationLink{
			Href:  baseUrl.String(),
			Label: "Last",
		})
	}
	results.PaginationLinks = links
}
