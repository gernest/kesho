package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccounts(t *testing.T) {
	astore := NewStorage("account)test.db", 0600)
	defer astore.DeleteDatabase()

	users := []string{"geofrey", "ernest", "gernest"}
	aBucket := "Account"

	Convey("Testing Accounts", t, func() {
		Convey("Creating new Accounts", func() {
			for _, v := range users {
				us := NewAccount(aBucket, astore)
				us.UserName = v
				err := us.Save()

				So(err, ShouldBeNil)
			}
		})
		Convey("Retrieve user from database", func() {
			for _, v := range users {
				us := NewAccount(aBucket, astore)
				us.UserName = v

				err := us.Get()

				So(err, ShouldBeNil)
				So(us.CreatedAt, ShouldNotBeNil) //TODO: add meaningful time comparizon for the field
			}
		})
		Convey("Retrieve All Users", func() {
			Convey("When the bucket is present", func() {
				acc := NewAccount(aBucket, astore)
				usr, err := acc.GetAllUsers()

				So(err, ShouldBeNil)
				So(len(usr), ShouldEqual, len(users))
			})
			Convey("When the bucket is not present", func() {
				acc := NewAccount("shit", astore)
				usr, err := acc.GetAllUsers()

				So(err, ShouldNotBeNil)
				So(len(usr), ShouldEqual, 0)
			})

		})
		Convey("Deleting an account", func() {
			us1 := NewAccount(aBucket, astore)
			us2 := NewAccount(aBucket, astore)

			us1.UserName = "gernest"
			err := us1.Delete()

			So(err, ShouldBeNil)

			us2.UserName = "gernest"
			err = us2.Get()

			So(err, ShouldNotBeNil)

		})
		Convey("Updating Account details", func() {
			us1 := NewAccount(aBucket, astore)
			us2 := NewAccount(aBucket, astore)

			us1.UserName = "geofrey"
			us1.BlogTitle = "yay"
			err := us1.Update()

			So(err, ShouldBeNil)

			us2.UserName = "geofrey"
			err = us2.Get()

			So(err, ShouldBeNil)

			So(us2.BlogTitle, ShouldEqual, us1.BlogTitle)

		})
	})

}
