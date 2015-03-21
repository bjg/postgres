package types

import (
	"bytes"
	"log"
	"strconv"
	"strings"
)

type Tuple struct {
	Key   string
	Value interface{}
}

type TupleList []Tuple

func (tl *TupleList) Scan(src interface{}) error {
	tuples := []Tuple{}
	buf := src.([]byte)
	buf = buf[1:]
	buf[len(buf)-1] = ','
	//log.Println(string(buf))
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
				var v interface{}
				if len(vals) > 2 {
					v = strings.Join(vals[1:], ",") // string(s)
				} else {
					// End of current tuple
					if vals[1] != "" {
						if integer, err := strconv.ParseInt(vals[1], 0, 64); err == nil {
							v = integer
						} else if float, err := strconv.ParseFloat(vals[1], 64); err == nil {
							v = float
						} else if boolean, err := strconv.ParseBool(vals[1]); err == nil {
							v = boolean
						} else {
							v = vals[1] // string
						}
					} else {
						v = ""
					}
				}
				t := Tuple{vals[0], v}
				tuples = append(tuples, t)
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
				// Squash double inverted commas
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
