package types

import (
	"fmt"
	"strconv"
	"strings"
)

type Point struct {
	lat, lng float64
}

func (p *Point) MarshalJSON() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Point) UnmarshalJSON(b []byte) error {
	var err error
	// Need to strip delimiting quote marks and parentheses
	s := string(b[2 : len(b)-2])
	parts := strings.Split(s, ",")
	if p.lat, err = strconv.ParseFloat(parts[0], 64); err == nil {
		p.lng, err = strconv.ParseFloat(parts[1], 64)
	}
	return err
}

func (p Point) String() string {
	return fmt.Sprintf(`(%v,%v)`, p.lat, p.lng)
}
