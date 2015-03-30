package main

import "testing"

func TestAccount_All(t *testing.T) {
	users := []struct {
		UserName string
		Password string
	}{
		{
			UserName: "gernest",
			Password: "gernestAHA",
		},
		{
			UserName: "ISIS",
			Password: "FUCkYOU",
		},
	}
	store := NewStore("accounts_test.db", 0600, nil)
	defer store.DeleteDatabase()
	var accBucket = "Accounts"

	for _, v := range users {
		usr := NewAccount(accBucket, store)
		usr.UserName = v.UserName
		usr.Password = v.Password
		usr.ConfirmPassword = v.Password

		err := usr.CreateUser()
		if err != nil {
			t.Error(err)
		}
	}

	acc := NewAccount(accBucket, store)

	for _, v := range users {
		usr, err := acc.GetUser(v.UserName)
		if err != nil {
			t.Error(err)
		}
		if !usr.IsValidPassword(v.Password) {
			t.Errorf("Expected passwords to match")
		}

		usr.Template = "kesho"
		err = usr.Update()
		if err != nil {
			t.Error(err)
		}
		us, _ := acc.GetUser(v.UserName)
		if us.Template != usr.Template {
			t.Errorf("Expected %s got %s", usr.Template, us.Template)
		}

	}
}
