package dto

import "encoding/json"

// OptionalUint razlikuje tri stanja JSON polja:
// - omitted  → Present=false
// - null     → Present=true, Value=nil
// - broj     → Present=true, Value=&n
type OptionalUint struct {
	Present bool
	Value   *uint
}

func (o *OptionalUint) UnmarshalJSON(data []byte) error {
	o.Present = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var value uint
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
}
