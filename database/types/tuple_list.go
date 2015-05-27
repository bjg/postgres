package types

import (
	"bytes"
	"strconv"
	"strings"
)

/*
 ("(id,4b663084-3f6c-41c2-821a-3e36c924ccbf)","(title,Test)"),("(id,113426a9-6f32-4947-9d6b-ac195bd8c460)","(title,Default)")
*/
type Tuple map[string]interface{}

func (t *Tuple) Scan(src interface{}) error {
	tp := scan(src.([]byte))
	*t = (*tp)[0]
	return nil
}

type TupleList []Tuple

func (tl *TupleList) Scan(src interface{}) error {
	tp := scan(src.([]byte))
	*tl = *tp
	return nil
}

func scan(buf []byte) *[]Tuple {
	tuples := []Tuple{}
	if len(buf) > 0 {
		buf = buf[1:]
		const (
			StartList = iota
			InTuple
			FinishTuple
		)
		state := StartList
		vals := []string{}
		current := make(Tuple)
		var elem *bytes.Buffer
		for i := 0; i < len(buf); {
			if state == StartList {
				if bytes.HasPrefix(buf[i:], []byte(`"(`)) {
					//log.Println(`Found "(`)
					state = InTuple
					elem = bytes.NewBuffer([]byte{})
					i += 2
				}
			} else if state == InTuple {
				if bytes.HasPrefix(buf[i:], []byte(`)"`)) {
					//log.Println(`Found )"`)
					state = FinishTuple
					// Value parsed
					vals = append(vals, elem.String())
					i += 2
				} else if buf[i] == ',' {
					// Key parsed
					vals = append(vals, elem.String())
					elem = bytes.NewBuffer([]byte{})
					i++
				} else {
					// Key or value byye
					elem.WriteByte(buf[i])
					i++
				}
			} else if state == FinishTuple {
				// We should have two vals elements
				var v interface{}
				if integer, err := strconv.ParseInt(vals[1], 0, 64); err == nil {
					v = integer
				} else if float, err := strconv.ParseFloat(vals[1], 64); err == nil {
					v = float
				} else if boolean, err := strconv.ParseBool(vals[1]); err == nil {
					v = boolean
				} else {
					v = strings.TrimSpace(strings.Replace(vals[1], `"`, ``, -1)) // string
				}
				current[vals[0]] = v
				vals = []string{}
				if tuples != nil && bytes.HasPrefix(buf[i:], []byte(`),(`)) {
					//log.Println(`Found ),(`)
					tuples = append(tuples, current)
					current = make(Tuple)
					state = StartList
					i += 3
				} else if i == len(buf)-1 {
					//log.Println(`End of buffer`)
					tuples = append(tuples, current)
					break
				} else {
					state = StartList
					i++
				}
			}
		}
	}
	return &tuples
}
