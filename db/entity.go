package db

import (
	"encoding/json"
	"unique"
)

type ID unique.Handle[string]

func Mkid(in string) ID {
	return ID(unique.Make(in))
}

func (i ID) String() string {
	return unique.Handle[string](i).Value()
}

func (i *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*i = ID(unique.Make(s))

	return nil
}

func (i ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

type Entity interface {
	ID() ID
	Type() ID
	SetID(s ID)
}
