package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

type Result struct {
	es  []string
	err error
}

func NewResult(err error) *Result {
	return &Result{err: err}
}

func (r *Result) One(model interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(r.es) < 1 {
		return errors.New("Result.One(): Expecting at least one model for decoding")
	}
	r.err = json.NewDecoder(strings.NewReader(r.es[0])).Decode(model)
	return r.err
}

func (r *Result) Each(model Model, do func(interface{})) error {
	if r.err != nil {
		return r.err
	}
	for _, enc := range r.es {
		instance := model.GetInstance()
		r.err = json.NewDecoder(strings.NewReader(enc)).Decode(instance)
		do(instance)
	}
	return r.err
}

func (r *Result) Update(enc string) *Result {
	if r.err == nil {
		r.es = append(r.es, enc)
	}
	return r
}

func (r *Result) Json() []byte {
	if len(r.es) <= 1 {
		return []byte(r.es[0])
	}
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, enc := range r.es {
		buf.WriteString(enc)
		if i < len(r.es)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("[")
	return buf.Bytes()
}

func (r *Result) NumRows() int {
	return len(r.es)
}

func (r *Result) Error() error {
	return r.err
}
