package main

import (
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func clearDataStoreForTest() {
	dsStore.Range(func(key, _ interface{}) bool {
		dsStore.Delete(key)
		return true
	})
}

func TestDataStoreSnapshotPersistsAcrossReload(t *testing.T) {
	originalPath := dsStorePath
	originalNextID := atomic.LoadUint64(&dsNextID)
	var originalObjects []*dsObject
	dsStore.Range(func(_, value interface{}) bool {
		originalObjects = append(originalObjects, value.(*dsObject))
		return true
	})

	t.Cleanup(func() {
		clearDataStoreForTest()
		for _, object := range originalObjects {
			dsStore.Store(object.DataID, object)
		}
		atomic.StoreUint64(&dsNextID, originalNextID)
		dsStorePath = originalPath
	})

	clearDataStoreForTest()
	dsStorePath = filepath.Join(t.TempDir(), "persistent", "datastore.json")
	atomic.StoreUint64(&dsNextID, 42)

	created := time.Date(2026, time.October, 13, 12, 30, 0, 0, time.UTC)
	want := &dsObject{
		DataID:     41,
		OwnerPID:   123456,
		DataType:   2,
		MetaBinary: []byte{0x10, 0x20, 0x30},
		Tags:       []string{"persist-test", "us_001_ave"},
		Name:       "persistence-test",
		Created:    created,
		Updated:    created,
	}
	dsStore.Store(want.DataID, want)

	dsSave()
	clearDataStoreForTest()
	atomic.StoreUint64(&dsNextID, 1)

	dsLoad()

	value, ok := dsStore.Load(want.DataID)
	if !ok {
		t.Fatalf("expected DataStore object %d to be restored after reload", want.DataID)
	}
	got := value.(*dsObject)
	if got.OwnerPID != want.OwnerPID {
		t.Errorf("OwnerPID = %d, want %d", got.OwnerPID, want.OwnerPID)
	}
	if got.DataType != want.DataType {
		t.Errorf("DataType = %d, want %d", got.DataType, want.DataType)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if len(got.MetaBinary) != len(want.MetaBinary) {
		t.Fatalf("MetaBinary length = %d, want %d", len(got.MetaBinary), len(want.MetaBinary))
	}
	for i := range want.MetaBinary {
		if got.MetaBinary[i] != want.MetaBinary[i] {
			t.Errorf("MetaBinary[%d] = 0x%02x, want 0x%02x", i, got.MetaBinary[i], want.MetaBinary[i])
		}
	}
	if nextID := atomic.LoadUint64(&dsNextID); nextID != 42 {
		t.Errorf("next DataStore ID = %d, want 42", nextID)
	}
}

func TestDatastoreFilePathUsesConfiguredPath(t *testing.T) {
	want := filepath.Join(t.TempDir(), "datastore.json")
	t.Setenv("WSC_DATASTORE_PATH", want)
	if got := datastoreFilePath(); got != want {
		t.Fatalf("datastoreFilePath() = %q, want %q", got, want)
	}
}

func TestDatastoreFilePathDefaultsWhenUnset(t *testing.T) {
	t.Setenv("WSC_DATASTORE_PATH", "  ")
	if got := datastoreFilePath(); got != "datastore.json" {
		t.Fatalf("datastoreFilePath() = %q, want %q", got, "datastore.json")
	}
}
