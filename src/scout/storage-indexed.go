package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type StorageIndexed struct {
	cookie string
	db     *bolt.DB

	lock   sync.Mutex
	offset uint64

	*StorageBase
}

func NewStorageIndexed(basename string, cookie string) *StorageIndexed {
	db, err := bolt.Open(basename+".idx", 0666, nil)
	if err != nil {
		log.Fatal("Can't create index:", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("sessions"))
		return err
	})
	if err != nil {
		log.Fatal("Can't create bucket 'session':", err)
	}

	return &StorageIndexed{
		cookie: cookie,
		db:     db,

		StorageBase: NewStorageBase(basename),
	}
}

func (s *StorageIndexed) Save(pair *ReqResPair) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.StorageBase.Save(pair)
	if err != nil {
		return err
	}

	err = s.index(pair)
	if err != nil {
		return err
	}

	bytes, err := pair.AsJson() // TODO: double pair to json conversion (first was in s.StorageBase.Save)
	if err != nil {
		return err
	}
	s.offset += uint64(len(bytes) + 1) // 1 is for '\n' byte

	return nil
}

func (s *StorageIndexed) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	return s.StorageBase.Close()
}

func (s *StorageIndexed) index(pair *ReqResPair) error {
	if !pair.HasRequest {
		return nil // skip broken pairs (which contain only response)
	}

	sess, err := pair.Cookie(s.cookie)
	if err == http.ErrNoCookie {
		return nil // don't index requests without session cookie
	}

	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("sessions"))
		orig := b.Get([]byte(sess.Value))
		b.Put([]byte(sess.Value), append(orig, s.itob(s.offset)...))

		return nil
	})
}

func (s *StorageIndexed) itob(value uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, value)
	return b
}
