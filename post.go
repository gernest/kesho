package main

import (
	"encoding/json"
	"time"

	"github.com/gosimple/slug"
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
	rec := p.Account.Store.CreateDataRecord(p.Account.Bucket, p.Slug, data, p.Account.UserName, "posts")
	return rec.Error
}
func (p *Post) Update() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	rec := p.Account.Store.UpdateDataRecord(p.Account.Bucket, p.Slug, data, p.Account.UserName, "posts")
	return rec.Error
}

func (p *Post) Get() error {
	var key string
	if p.Slug != "" {
		key = p.Slug
	} else if p.Slug == "" && p.Title != "" {
		key = slug.Make(p.Title)
	}
	rec := p.Account.Store.GetDataRecord(p.Account.Bucket, key, p.Account.UserName, "posts")
	if rec.Error != nil {
		return rec.Error
	}
	return json.Unmarshal(rec.Data, p)
}
