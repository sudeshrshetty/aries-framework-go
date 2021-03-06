/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/stretchr/testify/assert"
)

func TestNewDID(t *testing.T) {
	storedDoc := genesisDoc()
	require.NotNil(t, storedDoc)
	peerDID, err := newDid(storedDoc)
	require.NoError(t, err)
	require.NotNil(t, peerDID)
	assert.Contains(t, peerDID, "did:peer:11")
}

func TestNewDIDError(t *testing.T) {
	storedDoc := &did.Doc{ID: "did:peer:11"}
	_, err := newDid(storedDoc)
	require.Error(t, err)
	assert.Equal(t, err.Error(), "the genesis version must include public keys and authentication")
}

func TestComputeHash(t *testing.T) {
	hash, err := computeHash([]byte("Test"))
	assert.Nil(t, err)
	assert.NotNil(t, hash)
}
func TestComputeHashError(t *testing.T) {
	hash, err := computeHash([]byte(""))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty bytes")
	assert.Nil(t, hash)
}
func TestValidateDid(t *testing.T) {
	peerDoc, err := peerDidDoc()
	require.NoError(t, err)
	require.NotNil(t, peerDoc)
	err = validateDID(peerDoc)
	require.NoError(t, err)
}
func TestValidateDIDError(t *testing.T) {
	peerDoc := invalidPeerDIDDoc()
	require.NotNil(t, peerDoc)
	err := validateDID(peerDoc)
	require.Error(t, err)
	require.Equal(t, "did doesnt follow matching regex", err.Error())
}
func TestValidateErrorHashString(t *testing.T) {
	peerDoc := &did.Doc{ID: "did:peer:11-479cbc07c3f991725836a3aa2a581ca2029198aa420b9d99bc0e131d9f3e2cbe"}
	err := validateDID(peerDoc)
	require.Error(t, err)
	require.Equal(t, "hash of the doc doesnt match the computed hash", err.Error())
}

func TestValidateDIDRegex(t *testing.T) {
	did1 := &did.Doc{ID: "did:peer:22"}
	err := validateDID(did1)
	require.Error(t, err)
	require.Equal(t, err.Error(), "did doesnt follow matching regex")

	did2 := &did.Doc{ID: "did:sidetree:22"}
	err = validateDID(did2)
	require.Error(t, err)
	require.Equal(t, err.Error(), "did doesnt follow matching regex")

	did3 := &did.Doc{ID: "did:peer:1-*&$*|||"}
	err = validateDID(did3)
	require.Error(t, err)
	require.Equal(t, err.Error(), "did doesnt follow matching regex")
}

// genesisDoc creates the doc without an id
func genesisDoc() *did.Doc {

	pk := []did.PublicKey{
		{
			ID:         "did:example:123456789abcdefghi#keys-1",
			Type:       "Secp256k1VerificationKey2018",
			Controller: "did:example:123456789abcdefghi",
			Value:      []byte(`"publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"`),
		},
		{
			ID:         "did:example:123456789abcdefghw#key2",
			Type:       "RsaVerificationKey2018",
			Controller: "did:example:123456789abcdefghw",
			Value:      []byte(`"publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAryQICCl6NZ5gDKrnSztO\n3Hy8PEUcuyvg/ikC+VcIo2SFFSf18a3IMYldIugqqqZCs4/4uVW3sbdLs/6PfgdX\n7O9D22ZiFWHPYA2k2N744MNiCD1UE+tJyllUhSblK48bn+v1oZHCM0nYQ2NqUkvS\nj+hwUU3RiWl7x3D2s9wSdNt7XUtW05a/FXehsPSiJfKvHJJnGOX0BgTvkLnkAOTd\nOrUZ/wK69Dzu4IvrN4vs9Nes8vbwPa/ddZEzGR0cQMt0JBkhk9kU/qwqUseP1QRJ\n5I1jR4g8aYPL/ke9K35PxZWuDp3U0UPAZ3PjFAh+5T+fc7gzCs9dPzSHloruU+gl\nFQIDAQAB\n-----END PUBLIC KEY-----"`),
		},
	}
	auth := []did.VerificationMethod{
		{
			PublicKey: did.PublicKey{
				ID:         "did:example:123456789abcdefghs#key3",
				Type:       "RsaVerificationKey2018",
				Controller: "did:example:123456789abcdefghs",
				Value:      []byte(`"publicKeyHex": "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"`),
			},
		},
	}

	doc := &did.Doc{
		Context:        []string{"https://w3id.org/did/v1", "https://w3id.org/did/v2"},
		PublicKey:      pk,
		Authentication: auth,
		Created:        &time.Time{},
	}
	return &did.Doc{Context: doc.Context, PublicKey: doc.PublicKey, Authentication: doc.Authentication,
		Created: doc.Created}
}

func peerDidDoc() (*did.Doc, error) {
	doc := genesisDoc()
	did, err := newDid(doc)
	if err != nil {
		return nil, err
	}
	doc.ID = did
	return doc, nil
}

func invalidPeerDIDDoc() *did.Doc {
	doc := genesisDoc()
	doc.ID = "did:peer:11-"

	return doc
}
