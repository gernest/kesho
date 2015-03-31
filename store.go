package main

import (
	"bytes"
	"errors"
	"os"

	"github.com/boltdb/bolt"
)

type Store struct {
	Data     []byte
	DataList map[string][]byte
	Error    error
	Name     string
	Options  *bolt.Options
	Perm     os.FileMode
	db       *bolt.DB
}

func NewStore(name string, perm os.FileMode, options *bolt.Options) *Store {
	s := &Store{
		Name:    name,
		Perm:    perm,
		Options: options,
		Data:    nil,
	}
	return s.Open()
}

func (s *Store) create(bucket string, key string, value []byte) *Store {
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), value)
	})
	if s.Error == nil {
		s.Data = value
	}
	return s
}

func (s *Store) CreateRecord(bucket string, key string, value []byte, buckets ...string) *Store {
	return s.createRecord(bucket, key, value, buckets...)
}

func (s *Store) createRecord(bucket string, key string, value []byte, buckets ...string) *Store {
	if len(buckets) == 0 {
		return s.create(bucket, key, value)
	}
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		var prev *bolt.Bucket
		prev = b
		for i := 0; i < len(buckets); i++ {
			curr, err := prev.CreateBucketIfNotExists([]byte(buckets[i]))
			if err != nil {
				break
			}
			if curr == nil {
				continue
			}
			prev = curr
		}

		err = prev.Put([]byte(key), value)
		if err != nil {
			return err
		}

		rst := prev.Get([]byte(key))
		if rst != nil {
			s.Data = make([]byte, len((rst)))
			copy(s.Data, rst)
		}
		return nil
	})
	return s
}

func (s *Store) GetRecord(bucket, key string, buckets ...string) *Store {
	return s.getRecord(bucket, key, buckets...)
}

func (s *Store) getRecord(bucket, key string, buckets ...string) *Store {
	var uerr error = nil
	if len(buckets) == 0 {
		return s.get(bucket, key)
	}
	s.Error = s.db.View(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(buckets); i++ {
			curr := prev.Bucket([]byte(buckets[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + buckets[i] + "Not found")
				break
			}
			prev = curr
		}
		if uerr != nil {
			return uerr
		}

		rst := prev.Get([]byte(key))
		if rst == nil {
			return bolt.ErrBucketNotFound
		}
		s.Data = make([]byte, len(rst))
		copy(s.Data, rst)
		return nil
	})
	return s
}

func (s *Store) get(bucket string, key string) *Store {
	result := new(bytes.Buffer)
	s.Error = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		res := b.Get([]byte(key))
		if res != nil {
			_, err := result.Write(res)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if s.Error == nil {
		s.Data = result.Bytes()
	}
	return s
}

func (s *Store) PutRecord(bucket, key string, value []byte, buckets ...string) *Store {
	var uerr error
	if len(buckets) == 0 {
		return s.put(bucket, key, value)
	}
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(buckets); i++ {
			curr := prev.Bucket([]byte(buckets[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + buckets[i] + "Not found")
				break
			}
			prev = curr
		}
		if uerr != nil {
			return uerr
		}
		return prev.Put([]byte(key), value)
	})
	if s.Error == nil {
		s.Data = value
	}
	return s
}
func (s *Store) put(bucket string, key string, value []byte) *Store {

	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), value)
	})
	if s.Error == nil {
		s.Data = value
	}
	return s
}

// Size returns the size of the Store.Data field.
// if the field is nil, it will return 0
func (s *Store) Size() int64 {
	if s.Data == nil {
		return 0
	}
	return int64(len(s.Data))
}

func (s *Store) DeleteDatabase() error {
	path := s.db.Path()
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func (s *Store) RemoveRecord(bucket, key string, buckets ...string) *Store {
	return s.removeRecord(bucket, key, buckets...)
}

func (s *Store) removeRecord(bucket, key string, buckets ...string) *Store {
	var uerr error = nil
	if len(buckets) == 0 {
		return s.delete(bucket, key)

	}
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(buckets); i++ {
			curr := prev.Bucket([]byte(buckets[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + buckets[i] + "Not found")
				break
			}
			prev = curr
		}
		if uerr != nil {
			return uerr
		}
		perr := prev.Delete([]byte(key))
		if perr != nil {
			s.Data = nil
			return perr
		}
		s.Data = []byte(key)
		return nil

	})
	return s
}

func (s *Store) delete(bucket, key string) *Store {
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.Delete: Bucket" + bucket + " Not found")
		}
		return b.Delete([]byte(key))
	})
	return s
}

func (s *Store) GetAll(bucket string, buckets ...string) *Store {
	var uerr error = nil
	s.DataList = make(map[string][]byte)
	if len(buckets) == 0 {
		err := s.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return bolt.ErrBucketNotFound
			}
			return b.ForEach(func(k, v []byte) error {
				s.DataList[string(k)] = v
				return nil
			})
		})

		s.Error = err
		return s
	}

	s.Error = s.db.View(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(buckets); i++ {
			curr := prev.Bucket([]byte(buckets[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + buckets[i] + "Not found")
				break
			}
			prev = curr
		}

		if uerr != nil {
			return uerr
		}
		rerr := prev.ForEach(func(k, v []byte) error {
			s.DataList[string(k)] = v
			return nil
		})
		if rerr != nil {
			return rerr
		}
		return nil
	})
	return s
}

func (s *Store) Open() *Store {
	db, err := bolt.Open(s.Name, s.Perm, s.Options)
	s.db = db
	s.Error = err
	return s
}

func (s *Store) Close() error {
	return s.db.Close()
}
