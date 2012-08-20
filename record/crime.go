package record

import (
	"fmt"
	"strings"
)

type Offence struct {
	Id          string
	Category    string
	SubCategory string
	Desc        string
	Defendants  []*Person
	Verdict     *Verdict
}

type VerdictType string

const (
	GUILTY     VerdictType = "guilty"
	NOT_GUILTY             = "not guilty"
)

type Verdict struct {
	Id      string
	Verdict VerdictType
	Desc    string
}

func (v *Verdict) SetType(vtype string) {
	switch vtype {
	case "guilty":
		v.Verdict = GUILTY
	case "notGuilty":
		v.Verdict = NOT_GUILTY
	}
}

func (v *Verdict) String() string {
	return fmt.Sprintf("%v", *v)
}

func (v *Verdict) CssClass() string {
	return strings.Replace(string(v.Verdict), " ", "-", -1)
}