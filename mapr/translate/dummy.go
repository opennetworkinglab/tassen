package translate

import (
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

// A dummy translator for testing purposes only.
type dummyTranslator struct {
}

func NewDummyTranslator() Translator {
	return &dummyTranslator{}
}

// Returns the same input request.
func (d dummyTranslator) Translate(u *p4v1.Update) ([]*p4v1.Update, error) {
	return []*p4v1.Update{u}, nil
}

func (d dummyTranslator) ApplyUpdate(*p4v1.Update, []*p4v1.Update) error {
	return nil
}
