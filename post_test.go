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
	posts := []struct {
		Title, Body string
	}{
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
		err := usr.StampAndSave()
		if err != nil {
			t.Error(err)
		}
	}

	for _, usr := range users {

		for _, v := range posts {
			post := new(Post)
			post.Title = v.Title
			post.Body = v.Body
			post.Account = usr

			err := post.Create()
			if err != nil {
				t.Error(err)
			}
		}

	}

	for _, usr := range users {
		for _, v := range posts {
			post := new(Post)
			post.Title = v.Title
			post.Account = usr

			err := post.Get()
			if err != nil {
				t.Error(err)
			}
			if post.Body != v.Body {
				t.Errorf("Expected %s got %s", v.Body, post.Body)
			}
		}
	}
}
