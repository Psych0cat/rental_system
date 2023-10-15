package models

type AutoType struct {
	ID string `db:"id"`
}

func (a *AutoType) TableName() string {
	return "auto_type"
}
