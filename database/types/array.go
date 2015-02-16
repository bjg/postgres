package types

import (
	"bytes"
	"strings"
)

type Array struct {
	elems *[]string
}

func (a *Array) MarshalJSON() ([]byte, error) {
	return a.bytes(), nil
}

func (a *Array) UnmarshalJSON(b []byte) error {
	elems := []string{}
	for _, s := range strings.Split(string(b[1:len(b)-1]), ",") {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			elems = append(elems, s)
		}
	}
	a.elems = &elems
	return nil
}

func (a Array) String() string {
	return string(a.bytes())
}

func (a Array) Split() []string {
	if a.elems == nil {
		return []string{}
	}
	return *a.elems
}

func (a Array) bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("{")
	if a.elems != nil {
		for i, e := range *a.elems {
			buf.WriteString(e)
			if i < len(*a.elems)-1 {
				buf.WriteString(",")
			}
		}
	}
	buf.WriteString("}")
	return buf.Bytes()
}
