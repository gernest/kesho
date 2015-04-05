package main

import (
	"errors"
	"github.com/boltdb/bolt"
	"os"
)

type Storage struct {
	Data     []byte
	DataList map[string][]byte
	Error    error

	dbName string
	mode   os.FileMode
	db     *bolt.DB
}

type StorageFunc func(s Storage, bucket, key string, value []byte, nested ...string) Storage

func NewStorage(dbname string, mode os.FileMode) Storage {
	return Storage{
		dbName: dbname,
		mode:   mode,
	}
}

func (s Storage) CreateDataRecord(bucket, key string, value []byte, nested ...string) Storage {
	return s.execute(bucket, key, value, nested, createDataRecord)
}

func createDataRecord(s Storage, bucket, key string, value []byte, nested ...string) Storage {
	if len(nested) == 0 {
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
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		var prev *bolt.Bucket
		prev = b
		for i := 0; i < len(nested); i++ {
			curr, err := prev.CreateBucketIfNotExists([]byte(nested[i]))
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

func (s Storage) GetDataRecord(bucket, key string, nested ...string) Storage {
	return s.execute(bucket, key, nil, nested, getDataRecord)
}

func getDataRecord(s Storage, bucket, key string, value []byte, buckets ...string) Storage {
	var uerr error = nil
	if len(buckets) == 0 {
		s.Error = s.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return bolt.ErrBucketNotFound
			}
			res := b.Get([]byte(key))
			if res != nil {
				s.Data = make([]byte, len(res))
				copy(s.Data, res)
			}
			return nil
		})
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

func (s Storage) UpdateDataRecord(bucket, key string, value []byte, nested ...string) Storage {
	return s.execute(bucket, key, value, nested, updateDataRecord)
}

func updateDataRecord(s Storage, bucket, key string, value []byte, nested ...string) Storage {
	var uerr error
	if len(nested) == 0 {
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
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(nested); i++ {
			curr := prev.Bucket([]byte(nested[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + nested[i] + "Not found")
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

func (s Storage) GetAll(bucket string, nested ...string) Storage {
	return s.execute(bucket, "", nil, nested, getAll)
}

func getAll(s Storage, bucket, key string, value []byte, nested ...string) Storage {
	var uerr error = nil
	s.DataList = make(map[string][]byte)
	if len(nested) == 0 {
		err := s.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return bolt.ErrBucketNotFound
			}
			return b.ForEach(func(k, v []byte) error {
				dv := make([]byte, len(v))
				copy(dv, v)
				s.DataList[string(k)] = dv
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
		for i := 0; i < len(nested); i++ {
			curr := prev.Bucket([]byte(nested[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + nested[i] + "Not found")
				break
			}
			prev = curr
		}

		if uerr != nil {
			return uerr
		}
		rerr := prev.ForEach(func(k, v []byte) error {
			dv := make([]byte, len(v))
			copy(dv, v)
			s.DataList[string(k)] = dv
			return nil
		})
		if rerr != nil {
			return rerr
		}
		return nil
	})
	return s
}

func (s Storage) RemoveDataRecord(bucket, key string, nested ...string) Storage {
	return s.execute(bucket, key, nil, nested, removeDataRecord)
}

func removeDataRecord(s Storage, bucket, key string, value []byte, nested ...string) Storage {
	var uerr error = nil
	if len(nested) == 0 {
		s.Error = s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return errors.New("Store.Delete: Bucket" + bucket + " Not found")
			}
			return b.Delete([]byte(key))
		})
		return s
	}
	s.Error = s.db.Update(func(tx *bolt.Tx) error {
		var prev *bolt.Bucket
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("Store.GetRecord: bucket" + bucket + "not found")
		}
		prev = b
		for i := 0; i < len(nested); i++ {
			curr := prev.Bucket([]byte(nested[i]))
			if curr == nil {
				uerr = errors.New("Sore.GetRecord: Bucket " + nested[i] + "Not found")
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
func (s Storage) DeleteDatabase() error {
	return os.Remove(s.dbName)
}
func (s Storage) execute(bucket, key string, value []byte, nested []string, fn StorageFunc) Storage {
	s.db, s.Error = bolt.Open(s.dbName, s.mode, nil)
	if s.Error != nil {
		panic(s.Error)
	}
	defer s.db.Close()
	return fn(s, bucket, key, value, nested...)
}
