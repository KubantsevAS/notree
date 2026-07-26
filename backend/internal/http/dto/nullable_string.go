package dto

import "encoding/json"

type NullableString struct {
	Value *string
	IsSet bool
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.IsSet = true

	if string(data) == "null" {
		n.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}
