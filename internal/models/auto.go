package models

type Auto struct {
	ID           string `db:"id" sql:"type:VARCHAR(255)"`
	Type         string `db:"type" sql:"type:VARCHAR(255)"`
	Availability bool   `db:"availability" sql:"type:BOOLEAN"`
}

func (a *Auto) TableName() string {
	return "auto"
}
