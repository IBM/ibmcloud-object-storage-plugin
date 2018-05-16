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
	"github.com/stretchr/testify/assert"
	"testing"
)

type testType struct {
	T int `json:"t,string"`
}

func Test_DecodeBase64_Error(t *testing.T) {
	_, err := DecodeBase64("non-base64")
	assert.Error(t, err)
}

func Test_DecodeBase64_Positive(t *testing.T) {
	decoded, err := DecodeBase64("aGVsbG8gd29ybGQ=")
	if assert.NoError(t, err) {
		assert.Equal(t, "hello world", decoded)
	}
}

func Test_UnmarshalMap_Error(t *testing.T) {
	var v testType
	err := UnmarshalMap(&map[string]string{"t": "non-int-value"}, &v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot unmarshal")
	}
}

func Test_UnmarshalMap_Positive(t *testing.T) {
	var v testType
	err := UnmarshalMap(&map[string]string{"t": "5"}, &v)
	if assert.NoError(t, err) {
		assert.Equal(t, 5, v.T)
	}
}

func Test_MarshalToMap_MarshalError(t *testing.T) {
	type badType struct {
		F func()
	}
	_, err := MarshalToMap(&badType{F: func() {}})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot marshal object")
	}
}

func Test_MarshalToMap_UnmarshalError(t *testing.T) {
	_, err := MarshalToMap(&[]string{"a"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot unmarshal")
	}
}

func Test_MarshalToMap_NonStringValue(t *testing.T) {
	type badType struct {
		T int `json:"t"`
	}
	_, err := MarshalToMap(&badType{T: 1})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot convert value to string")
	}
}

func Test_MarshalToMap_Positive(t *testing.T) {
	v := testType{T: 5}
	m, err := MarshalToMap(&v)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"t": "5"}, m)
	}
}
