/*
 *
 * Copyright SecureKey Technologies Inc. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 * /
 *
 */

package connectionstore

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	mockstorage "github.com/hyperledger/aries-framework-go/pkg/internal/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
)

const (
	threadIDFmt  = "thID-%v"
	connIDFmt    = "connValue-%v"
	sampleErrMsg = "sample-error-message"
)

func TestNewConnectionReader(t *testing.T) {
	t.Run("create new connection reader", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{})
		require.NoError(t, err)
		require.NotNil(t, store)
		require.NotNil(t, store.TransientStore())
		require.NotNil(t, store.Store())
	})

	t.Run("create new connection reader failure due to transient store error", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{transientStoreError: fmt.Errorf(sampleErrMsg)})
		require.Error(t, err)
		require.Contains(t, err.Error(), sampleErrMsg)
		require.Nil(t, store)
	})

	t.Run("create new connection reader failure due to store error", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{storeError: fmt.Errorf(sampleErrMsg)})
		require.Error(t, err)
		require.Contains(t, err.Error(), sampleErrMsg)
		require.Nil(t, store)
	})
}

func TestConnectionReader_GetAndQueryConnectionRecord(t *testing.T) {
	const noOfItems = 12
	connectionIDS := make([]string, noOfItems)

	for i := 0; i < noOfItems; i++ {
		connectionIDS[i] = fmt.Sprintf(connIDFmt, i)
	}

	saveInStore := func(store storage.Store, ids []string) {
		for _, id := range ids {
			connRecBytes, err := json.Marshal(&ConnectionRecord{ConnectionID: id,
				ThreadID: fmt.Sprintf(threadIDFmt, id)})
			require.NoError(t, err)
			err = store.Put(GetConnectionKeyPrefix()(id), connRecBytes)
			require.NoError(t, err)
		}
	}

	t.Run("get connection record - from store", func(t *testing.T) {
		store, e := NewConnectionLookup(&mockProvider{})
		require.NoError(t, e)
		require.NotNil(t, store)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecord(connectionID)
			require.Error(t, err)
			require.Equal(t, err, storage.ErrDataNotFound)
			require.Nil(t, connection)
		}

		// prepare data
		saveInStore(store.Store(), connectionIDS)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecord(connectionID)
			require.NoError(t, err)
			require.NotNil(t, connection)
			require.Equal(t, connectionID, connection.ConnectionID)
			require.Equal(t, fmt.Sprintf(threadIDFmt, connectionID), connection.ThreadID)
		}

		records, e := store.QueryConnectionRecords()
		require.NoError(t, e)
		require.NotEmpty(t, records)
		require.Len(t, records, noOfItems)
	})

	t.Run("get connection record - from transient store", func(t *testing.T) {
		store, e := NewConnectionLookup(&mockProvider{})
		require.NoError(t, e)
		require.NotNil(t, store)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecord(connectionID)
			require.Error(t, err)
			require.Equal(t, err, storage.ErrDataNotFound)
			require.Nil(t, connection)
		}

		// prepare data
		saveInStore(store.TransientStore(), connectionIDS)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecord(connectionID)
			require.NoError(t, err)
			require.NotNil(t, connection)
			require.Equal(t, connectionID, connection.ConnectionID)
			require.Equal(t, fmt.Sprintf(threadIDFmt, connectionID), connection.ThreadID)
		}

		records, e := store.QueryConnectionRecords()
		require.NoError(t, e)
		require.NotEmpty(t, records)
		require.Len(t, records, noOfItems)
	})

	t.Run("get connection record - error scenario", func(t *testing.T) {
		provider := &mockProvider{}
		provider.store = &mockstorage.MockStore{ErrGet: fmt.Errorf(sampleErrMsg),
			Store: make(map[string][]byte)}
		store, err := NewConnectionLookup(provider)
		require.NoError(t, err)
		require.NotNil(t, store)

		// prepare data
		saveInStore(store.Store(), connectionIDS)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecord(connectionID)
			require.Error(t, err)
			require.Nil(t, connection)
			require.EqualError(t, err, sampleErrMsg)
		}
	})
}

func TestConnectionReader_GetConnectionRecordAtState(t *testing.T) {
	const state = "requested"

	const noOfItems = 12

	connectionIDS := make([]string, noOfItems)

	for i := 0; i < noOfItems; i++ {
		connectionIDS[i] = fmt.Sprintf(connIDFmt, i)
	}

	saveInStore := func(store storage.Store, ids []string) {
		for _, id := range ids {
			connRecBytes, err := json.Marshal(&ConnectionRecord{ConnectionID: id,
				ThreadID: fmt.Sprintf(threadIDFmt, id)})
			require.NoError(t, err)
			err = store.Put(GetConnectionStateKeyPrefix()(id, state), connRecBytes)
			require.NoError(t, err)
		}
	}

	t.Run("get connection record at state", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{})
		require.NoError(t, err)
		require.NotNil(t, store)

		// should fail since data doesn't exists
		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecordAtState(connectionID, state)
			require.Error(t, err)
			require.Equal(t, err, storage.ErrDataNotFound)
			require.Nil(t, connection)
		}

		// prepare data in store
		saveInStore(store.Store(), connectionIDS)

		// should fail since data doesn't exists in transient store
		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecordAtState(connectionID, state)
			require.Error(t, err)
			require.Equal(t, err, storage.ErrDataNotFound)
			require.Nil(t, connection)
		}

		// prepare data in transient store
		saveInStore(store.TransientStore(), connectionIDS)

		for _, connectionID := range connectionIDS {
			connection, err := store.GetConnectionRecordAtState(connectionID, state)
			require.NoError(t, err)
			require.NotNil(t, connection)
			require.Equal(t, connectionID, connection.ConnectionID)
			require.Equal(t, fmt.Sprintf(threadIDFmt, connectionID), connection.ThreadID)
		}
	})

	t.Run("get connection record at state - failure", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{})
		require.NoError(t, err)
		require.NotNil(t, store)

		connection, err := store.GetConnectionRecordAtState("sampleID", "")
		require.Error(t, err)
		require.EqualError(t, err, stateIDEmptyErr)
		require.Nil(t, connection)
	})
}

func TestConnectionReader_GetConnectionRecordByNSThreadID(t *testing.T) {
	const noOfItems = 12
	nsThreadIDs := make([]string, noOfItems)

	for i := 0; i < noOfItems; i++ {
		nsThreadIDs[i] = fmt.Sprintf(threadIDFmt, i)
	}

	saveInStore := func(store storage.Store, ids []string, skipConnection bool) {
		for _, id := range ids {
			connID := fmt.Sprintf(connIDFmt, id)
			connRecBytes, err := json.Marshal(&ConnectionRecord{ConnectionID: id,
				ThreadID: id})
			require.NoError(t, err)
			err = store.Put(id, []byte(connID))
			require.NoError(t, err)

			if !skipConnection {
				err = store.Put(GetConnectionKeyPrefix()(connID), connRecBytes)
				require.NoError(t, err)
			}
		}
	}

	t.Run("get connection record by NS thread ID", func(t *testing.T) {
		store, err := NewConnectionLookup(&mockProvider{})
		require.NoError(t, err)
		require.NotNil(t, store)

		// should fail since data doesn't exists
		for _, nsThreadID := range nsThreadIDs {
			connection, err := store.GetConnectionRecordByNSThreadID(nsThreadID)
			require.Error(t, err)
			require.Contains(t, err.Error(), storage.ErrDataNotFound.Error())
			require.Nil(t, connection)
		}

		// prepare data in store
		saveInStore(store.Store(), nsThreadIDs, false)

		// should fail since data doesn't exists in transient store
		for _, nsThreadID := range nsThreadIDs {
			connection, err := store.GetConnectionRecordByNSThreadID(nsThreadID)
			require.Error(t, err)
			require.Contains(t, err.Error(), storage.ErrDataNotFound.Error())
			require.Nil(t, connection)
		}

		// prepare only ns thread data in transient store
		// skip connection
		saveInStore(store.TransientStore(), nsThreadIDs, true)

		// should fail since data doesn't exists in transient store
		for _, nsThreadID := range nsThreadIDs {
			connection, err := store.GetConnectionRecordByNSThreadID(nsThreadID)
			require.Error(t, err)
			require.Contains(t, err.Error(), storage.ErrDataNotFound.Error())
			require.Nil(t, connection)
		}

		// prepare data in transient store
		saveInStore(store.TransientStore(), nsThreadIDs, false)

		// should fail since data doesn't exists in transient store
		for _, nsThreadID := range nsThreadIDs {
			connection, err := store.GetConnectionRecordByNSThreadID(nsThreadID)
			require.NoError(t, err)
			require.NotNil(t, connection)
			require.Equal(t, nsThreadID, connection.ThreadID)
		}
	})
}

func TestConnectionRecorder_PrepareConnectionRecord(t *testing.T) {
	t.Run(" prepare connection record  error", func(t *testing.T) {
		transientStore := &mockstorage.MockStore{Store: make(map[string][]byte)}
		record, err := NewConnectionLookup(&mockProvider{store: nil, transientStore: transientStore})
		require.NoError(t, err)
		require.NotNil(t, record)
		connRec, err := prepareConnectionRecord(nil)
		require.Contains(t, err.Error(), "prepare connection record")
		require.Nil(t, connRec)
	})
}

func TestConnectionRecorder_QueryConnectionRecord(t *testing.T) {
	t.Run("test query connection record", func(t *testing.T) {
		store := &mockstorage.MockStore{Store: make(map[string][]byte)}

		transientStore, err := mem.NewProvider().OpenStore(nameSpace)
		require.NoError(t, err)

		const (
			storeCount          = 5
			overlap             = 3
			transientStoreCount = 4
		)

		for i := 0; i < storeCount+overlap; i++ {
			val, jsonErr := json.Marshal(&ConnectionRecord{
				ConnectionID: string(i),
			})
			require.NoError(t, jsonErr)

			err = store.Put(fmt.Sprintf("%s_abc%d", connIDKeyPrefix, i), val)
			require.NoError(t, err)
		}
		for i := overlap; i < transientStoreCount+storeCount; i++ {
			val, jsonErr := json.Marshal(&ConnectionRecord{
				ConnectionID: string(i),
			})
			require.NoError(t, jsonErr)

			err = transientStore.Put(fmt.Sprintf("%s_abc%d", connIDKeyPrefix, i), val)
			require.NoError(t, err)
		}

		recorder, err := NewConnectionLookup(&mockProvider{store: store, transientStore: transientStore})
		require.NoError(t, err)
		require.NotNil(t, recorder)
		result, err := recorder.QueryConnectionRecords()
		require.NoError(t, err)
		require.Len(t, result, storeCount+transientStoreCount)
	})

	t.Run("test query connection record failure", func(t *testing.T) {
		store := &mockstorage.MockStore{Store: make(map[string][]byte)}
		err := store.Put(fmt.Sprintf("%s_abc123", connIDKeyPrefix), []byte("-----"))
		require.NoError(t, err)

		recorder, err := NewConnectionLookup(&mockProvider{store: store})
		require.NoError(t, err)
		require.NotNil(t, recorder)
		result, err := recorder.QueryConnectionRecords()
		require.Error(t, err)
		require.Empty(t, result)
	})
}

// mockProvider for connection recorder
type mockProvider struct {
	transientStoreError error
	storeError          error
	store               storage.Store
	transientStore      storage.Store
}

// TransientStorageProvider is mock transient storage provider for connection recorder
func (p *mockProvider) TransientStorageProvider() storage.Provider {
	if p.transientStoreError != nil {
		return &mockstorage.MockStoreProvider{ErrOpenStoreHandle: p.transientStoreError}
	}

	if p.transientStore != nil {
		return mockstorage.NewCustomMockStoreProvider(p.transientStore)
	}

	return mockstorage.NewMockStoreProvider()
}

// StorageProvider is mock storage provider for connection recorder
func (p *mockProvider) StorageProvider() storage.Provider {
	if p.storeError != nil {
		return &mockstorage.MockStoreProvider{ErrOpenStoreHandle: p.storeError}
	}

	if p.store != nil {
		return mockstorage.NewCustomMockStoreProvider(p.store)
	}

	return mockstorage.NewMockStoreProvider()
}