package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

type Post struct {
	Title     string    `json:"title"`
	Account   *Account  `json:"-"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at`
}

func (p *Post) Create() error {
	p.Author = p.Account.UserName
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return p.Account.Store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.Account.Bucket))
		if b == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.UserName + " Does not exixt")
		}
		ub := b.Bucket([]byte(p.Account.UserName))
		if ub == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.UserName + " Does not exixt")
		}
		nb, err := ub.CreateBucketIfNotExists([]byte("posts"))
		if err != nil {
			return err
		}
		return nb.Put([]byte(p.Title), data)
	})
}
func (p *Post) Save() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return p.Account.Store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.Account.UserName))
		if b == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.Bucket + " Does not exixt")
		}
		ub := b.Bucket([]byte(p.Account.UserName))
		if b == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.UserName + " Does not exixt")
		}
		nb := ub.Bucket([]byte("posts"))
		if nb == nil {
			return errors.New("kesho Post.Create: The bucket posts Does not exixt")
		}
		return nb.Put([]byte(p.Title), data)
	})
}

func (p *Post) Get() error {
	result := new(bytes.Buffer)
	err := p.Account.Store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.Account.Bucket))
		if b == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.Bucket + " Does not exixt")
		}

		ub := b.Bucket([]byte(p.Account.UserName))
		if b == nil {
			return errors.New("kesho Post.Create: The bucket " + p.Account.UserName + " Does not exixt")
		}
		nb := ub.Bucket([]byte("posts"))
		if nb == nil {
			return errors.New("kesho Post.Create: The bucket posts Does not exixt")
		}
		res := nb.Get([]byte(p.Title))
		if res != nil {
			read := bytes.NewReader(res)
			_, err := read.WriteTo(result)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return json.Unmarshal(result.Bytes(), p)
}
