package util

import (
	"encoding/json"
)

func PrettyPrint(v any) string {
	json, _ := json.MarshalIndent(v, "", "  ")
	return string(json)
}
