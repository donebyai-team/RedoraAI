package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

func valueAsJSON[T any](obj T, msg string) (driver.Value, error) {
	cnt, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal %s: %w", msg, err)
	}
	return cnt, nil
}

func scanFromJSON[T any](value interface{}, obj *T, msg string) error {
	if value == nil {
		obj = new(T)
		return nil
	}

	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return err
	}

	bs, ok := sv.([]byte)
	if !ok {
		return fmt.Errorf("could not convert data to byte array")
	}

	if err = json.Unmarshal(bs, &obj); err != nil {
		return fmt.Errorf("unable to unmarshal %s: %w", msg, err)
	}
	return nil
}
