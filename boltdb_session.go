package main

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type BStore struct {
	DB      *bolt.DB
	Codecs  []securecookie.Codec
	Options *sessions.Options

	defaultBStoreAge int
	bucketName       []byte
}

type sessionValue struct {
	Data    string    `json:"data"`
	Expires time.Time `json:"expires"`
}

func NewBStore(fileName string, bucketName string, defaultBStoreAge int, options *sessions.Options, keyPairs ...[]byte) (*BStore, error) {
	if defaultBStoreAge <= 0 {
		return nil, errors.New("defaultBStoreAge must be positive")
	}
	db, err := bolt.Open(fileName, 0640, nil)
	if err != nil {
		return nil, err
	}
	return NewBStoreFromDB(db, bucketName, defaultBStoreAge, options, keyPairs...)
}

func NewBStoreFromDB(db *bolt.DB, bucketName string, defaultBStoreAge int, options *sessions.Options, keyPairs ...[]byte) (*BStore, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &BStore{
		DB:               db,
		Codecs:           securecookie.CodecsFromPairs(keyPairs...),
		Options:          options,
		defaultBStoreAge: defaultBStoreAge,
		bucketName:       []byte(bucketName),
	}, nil
}

func (s *BStore) Close() {
	s.DB.Close()
}

func (s *BStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *BStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	session.Options = s.Options
	session.IsNew = true

	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

func (s *BStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.ID == "" {
		session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *BStore) Delete(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	options := *session.Options
	options.MaxAge = -1
	http.SetCookie(w, sessions.NewCookie(session.Name(), "", &options))
	for k := range session.Values {
		delete(session.Values, k)
	}
	err := s.DB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(s.bucketName).Delete([]byte(session.ID))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *BStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}
	value, err := json.Marshal(sessionValue{
		Data:    encoded,
		Expires: s.getExpires(session.Options.MaxAge),
	})
	return s.DB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(s.bucketName).Put([]byte(session.ID), value)
	})
}

func (s *BStore) load(session *sessions.Session) error {
	value := &sessionValue{}
	err := s.DB.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(s.bucketName).Get([]byte(session.ID))
		if data == nil {
			return nil
		}
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if value.Expires.Sub(time.Now()) < 0 {
		return errors.New("Session expired")
	}
	if err := securecookie.DecodeMulti(session.Name(), value.Data, &session.Values, s.Codecs...); err != nil {
		return err
	}
	return nil
}

func (s *BStore) getExpires(maxAge int) time.Time {
	if maxAge <= 0 {
		return time.Now().Add(time.Second * time.Duration(s.defaultBStoreAge))
	}
	return time.Now().Add(time.Second * time.Duration(maxAge))
}

func (s *BStore) DeleteExpired() error {
	return nil
}

func (s *BStore) DeleteExpiredPeriodic(period time.Duration) chan error {
	ticker := time.NewTicker(period)
	errChan := make(chan error)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.DeleteExpired(); err != nil {
					errChan <- err
				}
			}
		}
	}()
	return errChan
}
