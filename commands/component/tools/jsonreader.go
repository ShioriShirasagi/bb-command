package tools

import (
	"encoding/json"
	"os"
)

// ReadComponentDefineJSON コンポーネント定義JSONを読み込む
func ReadComponentDefineJSON(path string) (ComponentDefineJSON, error) {
	rawJSON, err := os.ReadFile(path)
	if err != nil {
		return ComponentDefineJSON{}, err
	}

	var componentDefineJSON ComponentDefineJSON
	if err := json.Unmarshal(rawJSON, &componentDefineJSON); err != nil {
		return ComponentDefineJSON{}, err
	}

	return componentDefineJSON, nil
}
