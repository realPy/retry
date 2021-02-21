package retrydb

import (
	"bytes"

	"gorm.io/gorm"
)

type RetryIt struct {
	ID   uint `gorm:"primaryKey"`
	Uuid string
	Data string
}

type RStoreGorm struct {
	db *gorm.DB
}

func NewRStoreWithGorm(db *gorm.DB) RStoreGorm {
	r := RStoreGorm{}
	r.db = db
	r.db.AutoMigrate(&RetryIt{})
	//register models

	return r
}

func (r RStoreGorm) Delete(uuid string) error {
	return r.db.Unscoped().Where("uuid = ?", uuid).Delete(&RetryIt{}).Error
}

func (r RStoreGorm) Store(uuid string, getData func() (*bytes.Buffer, error)) error {

	if b, err := getData(); err == nil {

		result := r.db.Model(&RetryIt{}).Where("uuid = ?", uuid).Update("data", b.String())

		if int(result.RowsAffected) != 1 {
			var retry = RetryIt{Uuid: uuid, Data: b.String()}

			return r.db.Create(&retry).Error
		}

	} else {
		return err
	}

	return nil
}

func (r RStoreGorm) ParseAll(parseData func(string, []byte) error) {
	var retryits []RetryIt

	if err := r.db.Find(&retryits).Error; err == nil {
		for _, retryit := range retryits {
			if err := parseData(retryit.Uuid, []byte(retryit.Data)); err != nil {
				r.Delete(retryit.Uuid)
			}

		}

	}

}
