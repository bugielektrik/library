package payment

import "encoding/json"

// interfaceToMap converts an interface{} to a map[string]interface{} using JSON marshaling.
func interfaceToMap(data interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}
