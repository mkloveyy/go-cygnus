package db

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONMap is a map field persist into db in json format and never be null, but use {} instead
type JSONMap map[string]interface{}

func (e JSONMap) Value() (driver.Value, error) {
	if e == nil {
		e = make(map[string]interface{})
	}

	j, err := json.Marshal(e)

	return string(j), err
}

func (e *JSONMap) Scan(value interface{}) error {
	if *e == nil {
		*e = make(map[string]interface{})
	}

	if value == nil {
		return nil
	}

	return json.Unmarshal(value.([]byte), e)
}
