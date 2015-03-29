package main

import (
	"testing"
)

func TestPost_All(t *testing.T) {
	users := []*Account{
		{
			UserName:        "gernest",
			Password:        "gernestAHA",
			ConfirmPassword: "gernestAHA",
		},
		{
			UserName:        "ISIS",
			ConfirmPassword: "FUCkYOU",
			Password:        "FUCkYOU",
		},
	}
	posts := []*Post{
		{
			Title: "Once upon a time in Tanzania",
			Body:  "He had a dream of saving his country",
		},
		{
			Title: "He tried and Tried and Tried",
			Body:  "Then one day His dream came true",
		},
	}

	postsDup := []*Post{
		{
			Title: "Once upon a time in Tanzania",
			Body:  "He had a dream of saving his country",
		},
		{
			Title: "He tried and Tried and Tried",
			Body:  "Then one day His dream came true",
		},
	}

	store := NewStore("accounts_test.db", 0600, nil)
	defer store.DeleteDatabase()

	var accBucket = "Accounts"

	// Create the accounts
	for _, usr := range users {
		usr.Store = store
		usr.Bucket = accBucket
		err := usr.CreateUser()
		if err != nil {
			t.Error(err)
		}
	}

	// Create the posts
	for _, usr := range users {
		for _, post := range posts {
			post.Account = usr
			err := post.Create()
			if err != nil {
				t.Error(err)
			}
		}
	}

	// Retrieve the posts
	for _, usr := range users {
		for _, post := range postsDup {
			post.Account = usr
			err := post.Get()
			if err != nil {
				t.Error(err)
			}
		}
	}
}
