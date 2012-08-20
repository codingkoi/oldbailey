package record

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var session *mgo.Session

func init() {
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
}

func FromCache(id string) *Record {
	var record Record
	err := session.DB("oldbailey").C("cases").FindId(id).One(&record)
	if err != nil && err.Error() == "not found" {
		return nil
	} else if err != nil {
		panic(err)
	}
	return &record
}

func (record *Record) Save() {
	selector := bson.M{"_id":record.Id}
	_, err := session.DB("oldbailey").C("cases").Upsert(selector, record)
	if err != nil {
		panic(err)
	}
}