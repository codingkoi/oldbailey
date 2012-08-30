package record

import (
	"fmt"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
	"github.com/moovweb/gokogiri/xpath"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unsafe"
)

const SUMMARY_LENGTH int = 200

type Record struct {
	Id          string "_id"
	Type        string
	RawText     []byte
	DisplayText string
	Persons     []Person
	Date        *time.Time
	Offences    []Offence
	Verdicts    []Verdict
	OfInterest  bool
	Note
}

func NewRecord(content []byte) (record *Record) {
	doc, err := gokogiri.ParseHtml([]byte(content))
	if err != nil {
		panic(err)
	}

	displayText := cleanUpContent(doc.String())
	record = &Record{RawText: content, DisplayText: displayText}
	dateStr := getInterp(doc.Root().NodePtr(), "date", doc)
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		record.Date = nil
	} else {
		record.Date = &date
	}

	xPath := xpath.NewXPath(doc.DocPtr())
	nodePtrs := xPath.Evaluate(doc.Root().NodePtr(),
		xpath.Compile("//div1"))

	node := xml.NewNode(nodePtrs[0], doc)
	record.Id = node.Attr("id")
	record.Type = node.Attr("type")

	record.processPersons(doc)
	record.processOffences(doc)
	record.processVerdicts(doc)
	record.processOffJoins(doc)
	return
}

/* Fetches a case record from the Old Bailey API */
func FetchRecord(id string) (record *Record) {
	// check the case first
	record = FromCache(id)
	if record != nil {
		return
	}

	resp, err := http.Get("http://www.oldbaileyonline.org/obapi/text?div=" + id)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	record = NewRecord(content)
	record.Save()
	return
}

func (record *Record) Summary() string {
	if len(record.DisplayText) > SUMMARY_LENGTH {
		return stripHtml(record.DisplayText[:SUMMARY_LENGTH] + "...")
	}
	return stripHtml(record.DisplayText)
}

func (record *Record) DisplayDate() string {
	if record.Date != nil {
		return fmt.Sprintf("%d %s, %d", record.Date.Day(), record.Date.Month(),
			record.Date.Year())
	}
	return "Date Unknown"
}

func (record *Record) Victims() (victims []Person) {
	victims = record.findPersonsByType(VICTIM)
	return
}

func (record *Record) Defendants() (defs []Person) {
	defs = record.findPersonsByType(DEFENDANT)
	return
}

func (record *Record) findPersonsByType(ptype PersonType) (ps []Person) {
	for _, person := range record.Persons {
		if person.Type == ptype {
			if !containsPerson(&ps, &person) {
				ps = append(ps, person)
			}
		}
	}
	return
}

func containsPerson(persons *[]Person, p *Person) bool {
	for _, person := range *persons {
		if duplicatePerson(&person, p) {
			return true
		}
	}
	return false
}

func duplicatePerson(p1 *Person, p2 *Person) bool {
	if p1.GivenName == p2.GivenName && p1.Surname == p2.Surname {
		return true
	}
	return false
}

func (record *Record) processPersons(doc *html.HtmlDocument) {
	xPath := xpath.NewXPath(doc.DocPtr())
	personPtrs := xPath.Evaluate(doc.Root().NodePtr(),
		xpath.Compile("//persname"))
	persons := make([]Person, len(personPtrs))

	for i, nodePtr := range personPtrs {
		node := xml.NewNode(nodePtr, doc)
		person := Person{}
		person.Id = node.Attr("id")
		person.GivenName = getInterp(nodePtr, "given", doc)
		person.Surname = getInterp(nodePtr, "surname", doc)
		person.SetType(node.Attr("type"))
		person.SetGender(getInterp(nodePtr, "gender", doc))
		persons[i] = person
	}
	record.Persons = persons
}

func (record *Record) processOffences(doc *html.HtmlDocument) {
	xPath := xpath.NewXPath(doc.DocPtr())
	offencePtrs := xPath.Evaluate(doc.Root().NodePtr(),
		xpath.Compile("//rs[@type='offenceDescription']"))
	offences := make([]Offence, len(offencePtrs))

	for i, nodePtr := range offencePtrs {
		node := xml.NewNode(nodePtr, doc)
		offence := Offence{}
		offence.Id = node.Attr("id")
		offence.Category = getInterp(nodePtr, "offenceCategory", doc)
		offence.SubCategory = getInterp(nodePtr, "offenceSubcategory", doc)
		offence.Desc = cleanUpContent(node.Content())
		offences[i] = offence
	}
	record.Offences = offences
}

func (record *Record) processVerdicts(doc *html.HtmlDocument) {
	xPath := xpath.NewXPath(doc.DocPtr())
	verdictPtrs := xPath.Evaluate(doc.Root().NodePtr(),
		xpath.Compile("//rs[@type='verdictDescription']"))
	verdicts := make([]Verdict, len(verdictPtrs))

	for i, nodePtr := range verdictPtrs {
		node := xml.NewNode(nodePtr, doc)
		verdict := Verdict{}
		verdict.Id = node.Attr("id")
		verdict.Desc = cleanUpContent(node.Content())
		verdict.SetType(getInterp(nodePtr, "verdictCategory", doc))
		verdicts[i] = verdict
	}
	record.Verdicts = verdicts
}

func (record *Record) processOffJoins(doc *html.HtmlDocument) {
	xPath := xpath.NewXPath(doc.DocPtr())
	// join the offence with the defendants and verdict
	joinPtrs := xPath.Evaluate(doc.Root().NodePtr(),
		xpath.Compile("//join[@result='criminalCharge']"))

	for _, nodePtr := range joinPtrs {
		node := xml.NewNode(nodePtr, doc)
		targets := strings.Split(node.Attr("targets"), " ")
		var personId, offId, verdictId string
		for _, targetId := range targets {
			if strings.Contains(targetId, "defend") {
				personId = targetId
			}
			if strings.Contains(targetId, "off") {
				offId = targetId
			}
			if strings.Contains(targetId, "verdict") {
				verdictId = targetId
			}
		}
		offence := record.findOffence(offId)
		if offence == nil {
			panic("couldn't find offence " + offId)
		}
		person := record.findPerson(personId)
		if person != nil {
			offence.Defendants = append(offence.Defendants, person)
		}
		verdict := record.findVerdict(verdictId)
		if verdict != nil {
			offence.Verdict = verdict
		}
	}
}

func (record *Record) findPerson(id string) (p *Person) {
	p = nil
	for i, person := range record.Persons {
		if person.Id == id {
			p = &record.Persons[i]
			return
		}
	}
	return
}

func (record *Record) findOffence(id string) (off *Offence) {
	off = nil
	for i, offence := range record.Offences {
		if offence.Id == id {
			off = &record.Offences[i]
			return
		}
	}
	return
}

func (record *Record) findVerdict(id string) (verdict *Verdict) {
	verdict = nil
	for i, verd := range record.Verdicts {
		if verd.Id == id {
			verdict = &record.Verdicts[i]
			return
		}
	}
	return
}

// get the value out of an <interp> tag
func getInterp(basePtr unsafe.Pointer, interpType string, doc *html.HtmlDocument) (value string) {
	xPath := xpath.NewXPath(doc.DocPtr())
	nodePtrs := xPath.Evaluate(basePtr, xpath.Compile(".//interp[@type='"+
		interpType+"']"))
	if len(nodePtrs) == 1 {
		node := xml.NewNode(nodePtrs[0], doc)
		value = node.Attr("value")
	}
	return
}

func cleanUpContent(content string) (result string) {
	// clear out all tags except <p>
	pattern, err := regexp.Compile("<pers(.|\n)+?>|<plac(.|\n)+?>|<[^p](.|\n)+?>")
	if err != nil {
		panic(err)
	}
	result = pattern.ReplaceAllString(content, " ")
	// minimiaze spaces in the output
	pattern, err = regexp.Compile("\\s+")
	if err != nil {
		panic(err)
	}
	result = pattern.ReplaceAllString(result, " ")
	// fix punctuation marks surrounded by space
	pattern, err = regexp.Compile("\\s([!-/:-@[-`{-~])")
	if err != nil {
		panic(err)
	}
	result = pattern.ReplaceAllString(result, "$1")
	return
}

func stripHtml(content string) (result string) {
	pattern, err := regexp.Compile("\\s*<(.|\n)+?>\\s*")
	if err != nil {
		panic(err)
	}
	result = pattern.ReplaceAllString(content, "")
	return
}
