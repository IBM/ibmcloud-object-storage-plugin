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
	"crypto/rand"
	"fmt"
	"io"
)

// Generator generates UUID strings
type Generator interface {
	// New generates a random UUID according to RFC 4122
	New() (string, error)
}

// ReaderGenerator generates UUID strings using an IO reader
type ReaderGenerator struct {
	// Reader is the entropy source for the UUID generator
	Reader io.Reader
}

// NewCryptoGenerator returns new cryptographic UUID generator
func NewCryptoGenerator() *ReaderGenerator {
	return &ReaderGenerator{Reader: rand.Reader}
}

// New generates a random UUID according to RFC 4122
func (u *ReaderGenerator) New() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(u.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
