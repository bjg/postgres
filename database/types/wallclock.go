package types

import "time"

const WC_FORMAT = "15:04:05+00"

type WallClock struct {
	time.Time
}

func (wc *WallClock) MarshalJSON() ([]byte, error) {
	return []byte(wc.String()), nil
}

func (wc *WallClock) UnmarshalJSON(b []byte) error {
	var err error
	// Need to strip delimiting quote marks
	s := string(b[1 : len(b)-1])
	// First try time with time zone format: 12:30:00+00
	wc.Time, err = time.Parse(WC_FORMAT, s)
	if err != nil {
		// Failing that, try kitchen format: 12:30PM
		wc.Time, err = time.Parse(time.Kitchen, s)
	}
	return err
}

func (wc WallClock) String() string {
	return wc.Format(WC_FORMAT)
}
