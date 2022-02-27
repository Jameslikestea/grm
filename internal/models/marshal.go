package models

import "encoding/json"

func Marshal(m Model) []byte {
	b, _ := json.Marshal(m)
	return b
}
