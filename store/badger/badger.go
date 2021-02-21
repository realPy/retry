package retrydb

import (
	"bytes"
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
)

type RStoreBadger struct {
	db *badger.DB
}

func NewRStoreBadger() RStoreBadger {
	r := RStoreBadger{}
	opts := badger.DefaultOptions("badger")
	opts.Logger = nil
	if db, err := badger.Open(opts); err != nil {
		fmt.Printf("error: %s", err)
	} else {
		r.db = db

	}
	return r
}

func (r RStoreBadger) Delete(uuid string) error {
	prefix_key := "srwl"
	key := fmt.Sprintf("%s_%s", prefix_key, uuid)

	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (r RStoreBadger) Store(uuid string, getData func() (*bytes.Buffer, error)) error {

	prefix_key := "srwl"
	key := fmt.Sprintf("%s_%s", prefix_key, uuid)

	return r.db.Update(func(txn *badger.Txn) error {
		if b, err := getData(); err == nil {
			return txn.Set([]byte(key), b.Bytes())
		} else {
			return err
		}
		return nil
	})

}

func (r RStoreBadger) ParseAll(parseData func(string, []byte) error) {

	r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("srwl_")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			if err := item.Value(func(v []byte) error {
				//fmt.Printf("key=%s, value=%s\n", k, v)
				if err := parseData(string(k), v); err != nil {
					return r.db.Update(func(txn *badger.Txn) error {
						return txn.Delete([]byte(k))
					})
				} else {
					return err
				}

			}); err != nil {
				return err
			}

		}
		return nil
	})

}
