/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// DecodeBase64 decodes a base64 string
func DecodeBase64(encoded string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// UnmarshalMap unmarshal a map[string]string to an interface (via JSON decoding)
func UnmarshalMap(m *map[string]string, v interface{}) error {
	jsonBytes, err := json.Marshal(*m)
	if err != nil {
		return fmt.Errorf("cannot marshal map: %v", err)
	}
	err = json.Unmarshal(jsonBytes, v)
	if err != nil {
		return fmt.Errorf("cannot unmarshal '%s': %v", string(jsonBytes), err)
	}
	return nil
}

// MarshalToMap converts an interface to map[string]string (via JSON encoding)
func MarshalToMap(v interface{}) (map[string]string, error) {
	var m map[string]interface{}

	jsonString, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal object: %v", err)
	}
	err = json.Unmarshal([]byte(jsonString), &m)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal '%s' to map: %v", jsonString, err)
	}

	res := make(map[string]string)

	for k, v := range m {
		stringVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("cannot convert value to string: %v", v)
		}
		res[k] = stringVal
	}
	return res, nil
}
