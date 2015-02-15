package types

import "time"

type Timestamp struct {
	time.Time
}

const TS_FORMAT = "2006-01-02 15:04:05.999999+00"

func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(ts.String()), nil
}

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	var err error
	// Need to strip delimiting quote marks
	s := string(b[1 : len(b)-1])
	ts.Time, err = time.Parse(TS_FORMAT, s)
	if err != nil {
		ts.Time, err = time.Parse(time.RFC3339Nano, s)
	}
	return err
}

func (ts Timestamp) String() string {
	return ts.Format(TS_FORMAT)
}
