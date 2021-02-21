package retrydb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type RStoreFS struct {
	path   string
	prefix string
}

func NewRStoreFS(path string, prefix string) RStoreFS {
	if err := os.MkdirAll(path, 0700); err != nil {
		panic(err)
	}
	r := RStoreFS{path: path, prefix: prefix}
	return r
}

func (r RStoreFS) Delete(uuid string) error {

	filename := fmt.Sprintf("%s%s", r.prefix, uuid)
	return os.Remove(path.Join(r.path, filename))

}

func (r RStoreFS) Store(uuid string, getData func() (*bytes.Buffer, error)) error {

	if b, err := getData(); err == nil {

		filename := fmt.Sprintf("%s%s", r.prefix, uuid)

		err := ioutil.WriteFile(path.Join(r.path, filename), b.Bytes(), 0600)

		return err

	} else {
		return err
	}

	return nil
}

func (r RStoreFS) ParseAll(parseData func(string, []byte) error) {

	filepath.Walk(r.path, func(pathname string, info os.FileInfo, err error) error {

		_, filename := path.Split(pathname)
		if strings.HasPrefix(filename, r.prefix) {
			s := strings.SplitN(filename, r.prefix, 2)
			if len(s) == 2 {
				if data, err := ioutil.ReadFile(pathname); err == nil {
					if err := parseData(s[1], data); err != nil {
						fmt.Printf("Want Delete file: %s but err %s\n", s[1], err.Error())
						//  r.Delete(s[1])
					}
				} else {
					fmt.Printf("Want Delete file: %s\n", s[1])
					//  r.Delete(s[1])
				}

			}

		}

		return nil
	})

	/*
	   if err := r.db.Find(&retryits).Error;err==nil {
	       for _,retryit:= range retryits {
	         if err:=parseData(retryit.Uuid,[]byte(retryit.Data));err!=nil {
	             r.Delete(retryit.Uuid)
	           }

	       }

	   }*/

}
