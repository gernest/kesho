package main

import (
	"encoding/json"
	"time"
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

	return p.Account.Store.CreateRecord(p.Account.Bucket, p.Title, data, p.Account.UserName, "posts").Error
}
func (p *Post) Update() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return p.Account.
		Store.PutRecord(p.Account.Bucket, p.Title, data, p.Account.UserName, "posts").Error
}

func (p *Post) Get() error {
	return json.Unmarshal(p.Account.Store.GetRecord(p.Account.Bucket, p.Title, p.Account.UserName, "posts").Data, p)
}
