package types

import (
	"database/sql/driver"
	"time"
)

type WallClock time.Time

func (wc WallClock) Value() (driver.Value, error) {
	return []byte(time.Time(wc).Format(time.Kitchen)), nil
}

func (wc *WallClock) Scan(src interface{}) error {
	*wc = WallClock(src.(time.Time))
	return nil
}

func (wc *WallClock) MarshalJSON() ([]byte, error) {
	v, err := wc.Value()
	return v.([]byte), err
}

func (wc *WallClock) UnmarshalJSON(b []byte) error {
	t, err := time.Parse(string(b[1:len(b)-1]), "15:04:00+00")
	*wc = WallClock(t)
	return err
}

func (wc WallClock) String() string {
	v, _ := wc.Value()
	return string(v.([]byte))
}
