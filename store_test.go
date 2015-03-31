package main

import (
	"testing"

	"fmt"
	"github.com/boltdb/bolt"
)

func TestStore_All(t *testing.T) {
	nests := [][]string{
		[]string{"one", "two"},
		[]string{"two", "three", "four"},
		[]string{"three", "four", "five", "six"},
		[]string{"four", "five", "six", "seven", "eight"},
	}
	data := []struct {
		key, value string
	}{
		{"key_1", "value_1"},
		{"key_2", "value_2"},
		{"key_3", "value_3"},
	}
	store := NewStore("test.db", 0600, nil)
	defer store.DeleteDatabase()

	defaultBucket := "storage"
	for _, v := range data {
		err := store.CreateRecord(defaultBucket, v.key, []byte(v.value)).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != v.value {
			t.Errorf("Expected %s got %s", v.value, store.Data)
		}
	}

	for k, v := range data {
		err := store.CreateRecord(defaultBucket, v.key, []byte(v.value), nests[k]...).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != v.value {
			t.Errorf("Expected %s got %s", v.value, store.Data)
		}
	}

	for _, v := range data {
		err := store.GetRecord(defaultBucket, v.key).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != v.value {
			t.Errorf("Expected %s got %s", v.value, store.Data)
		}
	}

	for k, v := range data {
		err := store.GetRecord(defaultBucket, v.key, nests[k]...).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != v.value {
			t.Errorf("Expected %s got %s", v.value, store.Data)
		}
	}
	for k, v := range data {
		x := fmt.Sprintf("%s %d", v.value, k)
		err := store.PutRecord(defaultBucket, v.key, []byte(x), nests[k]...).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != x {
			t.Errorf("Expected %s got %s", x, store.Data)
		}
	}
	for k, v := range data {
		x := fmt.Sprintf("%s %d", v.value, k)
		err := store.GetRecord(defaultBucket, v.key, nests[k]...).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != x {
			t.Errorf("Expected %s got %s", x, store.Data)
		}
	}

	err := store.CreateRecord(defaultBucket, "put", []byte("put")).Error
	if err != nil {
		t.Error(err)
	}
	err = store.PutRecord(defaultBucket, "put", []byte("put record")).Error
	if err != nil {
		t.Error(err)
	}
	err = store.GetRecord(defaultBucket, "put").Error
	if err != nil {
		t.Error(err)
	}
	if string(store.Data) != "put record" {
		t.Errorf("Expected put record got %s", store.Data)
	}
	store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(defaultBucket))
		b.CreateBucketIfNotExists([]byte("all_0ne"))
		b.CreateBucketIfNotExists([]byte("all_two"))
		b.CreateBucketIfNotExists([]byte("all_three"))
		return nil
	})
	err = store.GetAll(defaultBucket).Error
	if err != nil {
		t.Error(err)
	}
	err = store.GetAll(defaultBucket, "one").Error
	if err != nil {
		t.Error(err)
	}
	if len(store.DataList) != 1 {
		t.Errorf("Expected 1 got %d", len(store.DataList))
	}
	for k, v := range data {
		err := store.RemoveRecord(defaultBucket, v.key, nests[k]...).Error
		if err != nil {
			t.Error(err)
		}
		if string(store.Data) != v.key {
			t.Errorf("Expected %s got %s", v.value, store.Data)
		}
	}
}
