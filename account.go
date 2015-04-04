package main

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
)

type Account struct {
	// Storage
	Store  *Store `json:"-"`
	Bucket string `json:"-"`

	// Schema
	UserName        string    `json:"username`
	Password        string    `json:"password" `
	ConfirmPassword string    `json:"-"`
	BlogTitle       string    `json:"blog_title"`
	Theme           string    `json:"theme"`
	Template        string    `json:"template"`
	CreatedAt       time.Time `json:!created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Auth
	Email string `json:"email"`

	// OAuth2
	Oauth2Uid      string    `json:"oauth2_uid"`
	Oauth2Provider string    `json:"oauth2_provider"`
	Oauth2Token    string    `json:"oauth2_token"`
	Oauth2Refresh  string    `json:"aouth2_refresh"`
	Oauth2Expiry   time.Time `json:"oauth2_expiry"`

	// Confirm
	ConfirmToken string `json:"confirm_token"`
	Confirmed    bool   `json:"confirmed"`

	// Lock
	AttemptNumber int       `json:"attempt_number"`
	AttemptTime   time.Time `json:"attempt_time"`
	Locked        time.Time `json:"locked"`

	// Recover
	RecoverToken       string    `json:"recover_token"`
	RecoverTokenExpiry time.Time `json:"recover_token_expiry"`
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

func (acc *Account) create() error {
	return acc.Save()
}

func (acc *Account) StampAndSave() error {
	zero := new(time.Time)
	if acc.CreatedAt == *zero {
		acc.CreatedAt = time.Now()
	}
	acc.UpdatedAt = time.Now()
	return acc.Save()
}
func (acc *Account) CreateOauth(key, provider string) error {
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	return acc.Store.CreateRecord(acc.Bucket, key, data, "oaut", provider).Error
}

func (acc *Account) GetOauth(key, provider string) (*Account, error) {
	user := NewAccount(acc.Bucket, acc.Store)
	if acc.Store.GetRecord(acc.Bucket, key, "oauth", provider).Error != nil {
		return nil, acc.Store.Error
	}
	err := json.Unmarshal(acc.Store.Data, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (acc *Account) Update() error {
	acc.UpdatedAt = time.Now()
	return acc.Save()
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

func (acc *Account) GetAllUsers() (result []*Account, err error) {
	result = nil
	if acc.Store.GetAll(acc.Bucket).Error != nil {
		return nil, acc.Store.Error
	}
	for _, v := range acc.Store.DataList {
		user := NewAccount(acc.Bucket, acc.Store)
		err = json.Unmarshal(v, user)
		if err != nil {
			break
		}
		result = append(result, user)
	}

	return
}
