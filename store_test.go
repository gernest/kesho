package main

import (
	"bytes"
	"testing"
)

func TestStore_All(t *testing.T) {
	bucket := "store"
	data := []struct {
		key, val string
	}{
		{"One", "Moja"},
		{"two", "Mbili"},
	}

	nested := []string{"one", "two", "three", "four"}
	nestedData := struct {
		key, val, bucket string
	}{"I am", "One", "zero"}

	store := NewStore("stote_test.db", 0600, nil)
	defer store.DeleteDatabase()

	// Create
	for _, v := range data {
		store.Create(bucket, v.key, []byte(v.val))

		if !bytes.Equal(store.Data, []byte(v.val)) {
			t.Errorf("Expected %s to equal %s", store.Data, v.val)
		}
	}

	// Get
	for _, v := range data {
		store.Get(bucket, v.key)
		if !bytes.Equal(store.Data, []byte(v.val)) {
			t.Errorf("Expected %s to equal %s", store.Data, v.val)
		}
	}

	// Create a new Bucket
	newBuck := "mwanza"
	err := store.CreateBucket(newBuck)
	if err != nil {
		t.Error(err)
	}

	// Put some stuffs in the new bucket

	for _, v := range data {
		store.Put(newBuck, v.key, []byte(v.val))
		if !bytes.Equal(store.Data, []byte(v.val)) {
			t.Errorf("Expected %s to equal %s", store.Data, v.val)
		}
	}

	// Create a new record with nested buckets
	store.CreateRecord(nestedData.bucket, nestedData.key, []byte(nestedData.val), nested...)
	if store.Error != nil {
		t.Log(store.Error)
	}
	if string(store.Data) != nestedData.val {
		t.Errorf("Expected %s got %s", nestedData.val, store.Data)
	}

	// Retrieve A record with nested buckets
	store.gerRecord(nestedData.bucket, nestedData.key, nested...)
	if string(store.Data) != nestedData.val {
		t.Errorf("Expected %s got %s", nestedData.val, store.Data)
	}

	store.RemoveRecord(nestedData.bucket, nestedData.key, nested...)
	if store.Error != nil {
		t.Error(store.Error)
	}
}
