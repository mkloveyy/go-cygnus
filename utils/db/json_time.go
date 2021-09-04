package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

func (t *JSONTime) IsNull() bool {
	return t.Time.UnixNano() == time.Time{}.UnixNano()
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t *JSONTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	t.Time, err = time.Parse("2006-01-02 15:04:05", string(data))

	return
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	if t.IsNull() {
		return nil, nil
	}

	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}

	return fmt.Errorf("can not convert %v to timestamp", v)
}
