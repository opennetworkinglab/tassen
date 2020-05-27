package translators

import v1 "github.com/p4lang/p4runtime/go/p4/v1"

// A dummy translator for testing purposes only.
type Dummy struct {
}

// Returns the same input request.
func (d Dummy) Translate(request *v1.WriteRequest) (*v1.WriteRequest, error) {
	return request, nil
}
