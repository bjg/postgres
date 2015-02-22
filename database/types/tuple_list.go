package types

import (
	"bytes"
	"log"
)

type Tuple struct {
	Key   string
	Label string
	Value string
}

type TupleList []Tuple

func (tl *TupleList) Scan(src interface{}) error {
	tuples := []Tuple{}
	buf := src.([]byte)
	buf = buf[1:]
	buf[len(buf)-1] = ','
	const (
		Start = iota
		Open
	)
	state := Start
	vals := []string{}
	var value *bytes.Buffer
	parens := 0
	for i := 0; i < len(buf); {
		switch state {
		case Start:
			if buf[i] == ',' {
				// End of current tuple
				if len(vals[2]) > 1 {
					tuples = append(tuples, Tuple{vals[0], vals[1], vals[2]})
				}
				vals = []string{}
				i++
			} else if bytes.HasPrefix(buf[i:], []byte(`"(`)) {
				state = Open
				i += 2
				value = bytes.NewBuffer([]byte{})
			} else {
				log.Fatal(string(buf[i:]), i)
			}
		case Open:
			if bytes.HasPrefix(buf[i:], []byte(`""`)) {
				// Skip over double inverted commas
				i += 2
			} else if parens == 0 && bytes.HasPrefix(buf[i:], []byte(`)"`)) {
				// End of current value and tuple
				state = Start
				vals = append(vals, value.String())
				i += 2
			} else if parens == 0 && buf[i] == ',' {
				// End of current value
				vals = append(vals, value.String())
				value = bytes.NewBuffer([]byte{})
				i++
			} else {
				// Count value-internal parens
				if buf[i] == '(' {
					parens++
				} else if buf[i] == ')' {
					parens--
				}
				value.WriteByte(buf[i])
				i++
			}
		}
	}
	*tl = tuples
	return nil
}
