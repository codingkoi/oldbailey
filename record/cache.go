package record

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var session *mgo.Session

type CacheResults struct {
	Hits  []string
	Total int
	Start int
	Count int
}

func init() {
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
}

func CasesDB() *mgo.Collection {
	return session.DB("oldbailey").C("cases")
}

func FromCache(id string) *Record {
	var record Record
	err := CasesDB().FindId(id).One(&record)
	if err != nil && err.Error() == "not found" {
		return nil
	} else if err != nil {
		panic(err)
	}
	return &record
}

func (record *Record) Save() {
	selector := bson.M{"_id": record.Id}
	_, err := CasesDB().Upsert(selector, record)
	if err != nil {
		panic(err)
	}
}

func FetchSavedCases(subcat string, start, count int) (results CacheResults) {
	criteria := bson.M{
		"offences.subcategory": subcat,
		"ofinterest":           true,
	}
	query := CasesDB().Find(criteria).Sort("_id")
	results.Total, _ = query.Count()
	query.Skip(start)
	query.Limit(count)
	query.Select(bson.M{"_id": 1})

	iter := query.Iter()
	var result struct{ Id string "_id" }
	for iter.Next(&result) {
		results.Hits = append(results.Hits, result.Id)
	}
	if iter.Err() != nil {
		panic(iter.Err())
	}
	results.Start = start
	// at this point the query is limited to what we actually wanted
	// so query.Count is the count of what we got
	results.Count, _ = query.Count()
	return
}
