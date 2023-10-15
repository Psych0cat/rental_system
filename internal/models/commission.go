package models

type Commission struct {
	AutoType     string `db:"auto_type"`
	Type         string `db:"type"`
	Value        int    `db:"value"`
	MinThreshold int    `db:"min_threshold"`
}

func (a *Commission) TableName() string {
	return "commission"
}
