package record

import "fmt"

type PersonType string

const (
	VICTIM    PersonType = "victim"
	DEFENDANT            = "defendant"
)

type PersonGender string

const (
	MALE   PersonGender = "♂"
	FEMALE              = "♀"
)

type Person struct {
	Id        string
	Surname   string
	GivenName string
	Type      PersonType
	Gender    PersonGender
}

func (p *Person) SetType(ptype string) {
	switch ptype {
	case "victimName":
		p.Type = VICTIM
	case "defendantName":
		p.Type = DEFENDANT
	}
}

func (p *Person) SetGender(gender string) {
	switch gender {
	case "male":
		p.Gender = MALE
	case "female":
		p.Gender = FEMALE
	}
}

func (p *Person) String() string {
	return fmt.Sprintf("%v", *p)
}
