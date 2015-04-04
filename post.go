package main

import (
	"encoding/json"
	"github.com/gosimple/slug"
	"time"
)

type Post struct {
	Title     string    `json:"title"`
	Account   *Account  `json:"-"`
	Slug      string    `json:"slug"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at`
}

func (p *Post) Create() error {
	p.Author = p.Account.UserName
	p.Slug = slug.Make(p.Title)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return p.Account.Store.
		CreateRecord(p.Account.Bucket, p.Slug, data, p.Account.UserName, "posts").Error
}
func (p *Post) Update() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return p.Account.Store.
		PutRecord(p.Account.Bucket, p.Title, data, p.Account.UserName, "posts").Error
}

func (p *Post) Get() error {
	var key string
	if p.Slug != "" {
		key = p.Slug
	} else if p.Slug == "" && p.Title != "" {
		key = slug.Make(p.Title)
	}
	p.Account.Store.GetRecord(p.Account.Bucket, key, p.Account.UserName, "posts")
	if p.Account.Store.Error != nil {
		return p.Account.Store.Error
	}
	return json.Unmarshal(p.Account.Store.Data, p)
}
