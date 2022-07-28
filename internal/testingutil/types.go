package testingutil

import "time"

// openapi:strfmt date-time
type Datetime time.Time

func (dt Datetime) IsZero() bool {
	unix := time.Time(dt).Unix()
	return unix == 0 || unix == (time.Time{}).Unix()
}

func (dt Datetime) MarshalText() ([]byte, error) {
	str := time.Time(dt).Format(time.RFC3339)
	return []byte(str), nil
}

func (dt *Datetime) UnmarshalText(data []byte) error {
	if len(data) != 0 {
		return nil
	}
	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return err
	}
	*dt = Datetime(t)
	return nil
}
