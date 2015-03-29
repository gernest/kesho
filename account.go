package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

const (
	PASSWORD_HASHCOST = 8
)

type Account struct {
	// Storage
	Store  *Store `json:"-"`
	Bucket string `json:"-"`

	// Schema
	UserName        string    `json:"username"  valid:"Required;AlphaNumeric"`
	Password        string    `json:"password"  valid:"Required"`
	ConfirmPassword string    `json:"- valid:"Required"`
	BlogTitle       string    `json:"blog_title"`
	Theme           string    `json:"theme"`
	Template        string    `json:"template"`
	CreatedAt       time.Time `json:!created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewAccount(bucket string, db *Store) *Account {
	return &Account{Store: db, Bucket: bucket}
}

func (acc *Account) Save() error {
	return acc.saveUser()
}

func (acc *Account) saveUser() error {
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	return acc.Store.CreateRecord(acc.Bucket, acc.UserName, data, acc.UserName).Error
}

func (acc *Account) Get() error {
	if acc.Store.GetRecord(acc.Bucket, acc.UserName, acc.UserName).Error != nil {
		return acc.Store.Error
	}
	err := json.Unmarshal(acc.Store.Data, acc)
	if err != nil {
		return err
	}
	return nil
}

func (acc *Account) GetUser(name string) (*Account, error) {
	user := NewAccount(acc.Bucket, acc.Store)
	if acc.Store.GetRecord(acc.Bucket, name, name).Error != nil {
		return nil, acc.Store.Error
	}
	err := json.Unmarshal(acc.Store.Data, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (acc *Account) CreateUser() error {
	if acc.Password != acc.ConfirmPassword {
		return errors.New("kesho Account.CreateUser: Password is not equal to ConfirmPassword Field")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(acc.Password), PASSWORD_HASHCOST)
	if err != nil {
		return err
	}
	acc.CreatedAt = time.Now()
	acc.UpdatedAt = time.Now()
	acc.Password = string(h)

	return acc.create()
}
func (acc *Account) create() error {
	err := acc.Store.CreateBucket(acc.Bucket)
	if err != nil {
		return err
	}
	return acc.Save()
}

func (acc *Account) Update() error {
	acc.UpdatedAt = time.Now()
	return acc.Save()
}
func (acc *Account) IsUser() bool {
	return acc.hasUser(acc.UserName)
}

func (acc *Account) hasUser(name string) bool {
	if _, err := acc.GetUser(name); err != nil {
		return false
	}
	return true
}

func (acc *Account) login() error {
	pass := acc.Password
	err := acc.Get()
	if err != nil {
		return err
	}
	_, err = acc.validatePassword(pass)
	return err
}

func (acc *Account) IsValidPassword(pass string) bool {
	b, _ := acc.validatePassword(pass)
	return b
}

func (acc *Account) validatePassword(pass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(pass))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (acc *Account) DeleteUser() error {
	return acc.Delete(acc.UserName)
}

func (acc *Account) Delete(name string) error {
	return acc.deleteUser(name)
}

func (acc *Account) deleteUser(name string) error {
	return acc.Store.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
}
