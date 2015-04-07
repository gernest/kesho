package main

import (
	"encoding/json"
	"time"

	"errors"
)

type Account struct {
	// Storage
	Store  Storage `json:"-"`
	Bucket string  `json:"-"`

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

func NewAccount(bucket string, db Storage) *Account {
	return &Account{Store: db, Bucket: bucket}
}

func (acc *Account) Save() error {
	zero := new(time.Time)
	if acc.CreatedAt == *zero {
		acc.CreatedAt = time.Now()
	}
	acc.UpdatedAt = time.Now()
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	rec := acc.Store.CreateDataRecord(acc.Bucket, acc.UserName, data, acc.UserName)
	return rec.Error
}

func (acc *Account) Get() error {
	rec := acc.Store.GetDataRecord(acc.Bucket, acc.UserName, acc.UserName)
	if rec.Error != nil {
		return rec.Error
	} else if rec.Data == nil {
		return errors.New("No Record Found")
	}
	err := json.Unmarshal(rec.Data, acc)
	if err != nil {
		return err
	}
	return nil
}

func (acc *Account) CreateOauth(key, provider string) error {
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	rec := acc.Store.CreateDataRecord(acc.Bucket, key, data, "oaut", provider)
	return rec.Error
}

func (acc *Account) GetOauth(key, provider string) (*Account, error) {
	user := NewAccount(acc.Bucket, acc.Store)
	rec := acc.Store.GetDataRecord(acc.Bucket, key, "oauth", provider)
	if rec.Error != nil {
		return nil, rec.Error
	} else if rec.Data == nil {
		return nil, errors.New("No Such Record")
	}

	err := json.Unmarshal(rec.Data, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (acc *Account) Update() error {
	acc.UpdatedAt = time.Now()
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	rec := acc.Store.UpdateDataRecord(acc.Bucket, acc.UserName, data, acc.UserName)
	return rec.Error
}

func (acc *Account) Delete() error {
	rec := acc.Store.RemoveDataRecord(acc.Bucket, acc.UserName, acc.UserName)
	return rec.Error
}

func (acc *Account) GetAllUsers() (result []*Account, err error) {
	all := acc.Store.GetAll(acc.Bucket)
	if all.Error != nil {
		return nil, all.Error
	}
	for k, _ := range all.DataList {
		user := NewAccount(acc.Bucket, acc.Store)
		user.UserName = k
		err = user.Get()
		if err == nil {
			result = append(result, user)
		}
	}
	return
}
