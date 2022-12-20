package idgen

import "github.com/google/uuid"

type idgen struct{}

func New() *idgen {
	return &idgen{}
}

func (i *idgen) GenID() string {
	return uuid.New().String()
}
