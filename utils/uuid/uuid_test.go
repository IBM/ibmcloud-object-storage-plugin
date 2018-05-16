/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package uuid

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_New_ReadError(t *testing.T) {
	r := ReaderGenerator{Reader: bytes.NewReader(nil)}
	_, err := r.New()
	assert.Error(t, err)
}

func Test_CryptoGenerator_Positive(t *testing.T) {
	m := make(map[string]interface{})
	r := NewCryptoGenerator()
	for i := 0; i < 100; i++ {
		val, err := r.New()
		if !assert.NoError(t, err) {
			break
		}
		_, found := m[val]
		if !assert.False(t, found) {
			break
		}
		m[val] = nil
	}
}
