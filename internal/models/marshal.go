package models

import "encoding/json"

func Marshal(m Model) []byte {
	b, _ := json.Marshal(m)
	return b
}

func Unmarshal(c []byte, m Model) error {
	return json.Unmarshal(c, m)
}
