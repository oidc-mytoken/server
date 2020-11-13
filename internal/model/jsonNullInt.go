package model

import "encoding/json"

type JSONNullInt struct {
	Value int
	Valid bool
}

func (i *JSONNullInt) MarshalJSON() ([]byte, error) {
	j := i.Value
	if !i.Valid {
		j = -1
	}
	return json.Marshal(j)
}

func (i *JSONNullInt) UnmarshalJSON(data []byte) error {
	var j int
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	i.Value = j
	if j >= 0 {
		i.Valid = true
	}
	return nil
}
