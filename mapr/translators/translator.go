package translators

import (
	"github.com/p4lang/p4runtime/go/p4/v1"
)

// A P4Runtime translator.
type Translator interface {
	// Given a write request for the logical pipeline, returns a write request for the physical pipeline that would
	// produce a forwarding state that is equivalent to that of the logical pipeline after applying the given write
	// request. The returned request might be nil, signaling that the translation was successful but that should not
	// produce any change in the physical pipeline.
	Translate(*v1.WriteRequest) (*v1.WriteRequest, error)
}
