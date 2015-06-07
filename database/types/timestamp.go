package types

import (
	"database/sql/driver"
	"time"
)

type Timestamp time.Time

const format = "01/02/2006 15:04:00"

func (ts Timestamp) Value() (driver.Value, error) {
	return []byte(time.Time(ts).Format(format)), nil
}

func (ts *Timestamp) Scan(src interface{}) error {
	*ts = Timestamp(src.(time.Time))
	return nil
}

func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ts.String() + `"`), nil
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	// Need to strip delimiting quote marks
	s := string(b[1 : len(b)-1])
	t, err := time.Parse(format, s)
	*ts = Timestamp(t)
	return err
}

func (ts Timestamp) String() string {
	return time.Time(ts).Format(format)
}
