package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPost(t *testing.T) {

	astore := NewStorage("post_test.db", 0600)
	aBucket := "Accounts"
	defer astore.DeleteDatabase()

	users := []string{"geofrey", "ernest", "gernest"}

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

	Convey("Testing posts", t, func() {
		for _, v := range users {
			us := NewAccount(aBucket, astore)
			us.UserName = v
			err := us.Save()

			So(err, ShouldBeNil)
		}
		Convey("Creating Posts", func() {
			for _, v := range users {
				us := NewAccount(aBucket, astore)
				us.UserName = v
				err := us.Get()

				So(err, ShouldBeNil)
				for _, post := range posts {
					p := new(Post)
					p.Account = us
					p.Title = post.Title
					p.Body = post.Body

					perr := p.Create()

					So(perr, ShouldBeNil)

				}

			}
		})
		Convey("Retrieving posts", func() {
			for _, v := range users {
				us := NewAccount(aBucket, astore)
				us.UserName = v
				err := us.Get()

				So(err, ShouldBeNil)
				for _, post := range posts {
					p := new(Post)
					p.Account = us
					p.Title = post.Title

					perr := p.Get()

					So(perr, ShouldBeNil)
					So(p.Body, ShouldEqual, post.Body)
				}
			}
		})
		Convey("Updating posts", func() {
			for _, v := range users {
				us := NewAccount(aBucket, astore)
				us.UserName = v
				err := us.Get()

				So(err, ShouldBeNil)
				for _, post := range posts {
					p := new(Post)
					p2 := new(Post)

					p.Account = us
					p.Title = post.Title

					perr := p.Get()
					So(perr, ShouldBeNil)

					p.Body = post.Title

					p2.Account = us
					p2.Title = post.Title

					perr = p.Update()
					perr2 := p2.Get()

					So(perr, ShouldBeNil)
					So(perr2, ShouldBeNil)
					So(p2.Body, ShouldEqual, post.Title)
				}
			}
		})

	})

}
