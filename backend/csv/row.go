package csv

import (
	"time"
)

type BatchRow struct {
	FirstName   string    `validate:"required"`
	LastName    string    `validate:"required"`
	Phone       string    `validate:"required"`
	DueDate     time.Time `validate:"required"`
	ProductType string    `validate:"required"`
}

func NewBatchRow() *BatchRow {
	return &BatchRow{}
}

func (r *BatchRow) SetFirstName(firstName string) {
	r.FirstName = firstName
}

func (r *BatchRow) SetLastName(lastName string) {
	r.LastName = lastName
}

func (r *BatchRow) SetPhone(phone string) {
	r.Phone = phone
}

func (r *BatchRow) SetDueDate(dueDate time.Time) {
	r.DueDate = dueDate
}

func (r *BatchRow) SetProductType(productType string) {
	r.ProductType = productType
}
