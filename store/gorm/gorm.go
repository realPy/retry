package retrydb

import (
	"bytes"

	"gorm.io/gorm"
)

//Retry Retry struct gorm representation
type Retry struct {
	ID   uint `gorm:"primaryKey"`
	UUID string
	Data string
}

//RStoreGorm RStoreGorm struct
type RStoreGorm struct {
	db *gorm.DB
}

//NewRStoreWithGorm return a NewRStoreWithGorm instance
func NewRStoreWithGorm(db *gorm.DB) RStoreGorm {
	r := RStoreGorm{}
	r.db = db
	r.db.AutoMigrate(&Retry{})
	//register models

	return r
}

//Delete Delete the element in store
func (r RStoreGorm) Delete(UUID string) error {
	return r.db.Unscoped().Where("uuid = ?", UUID).Delete(&Retry{}).Error
}

//Store store the element in store
func (r RStoreGorm) Store(UUID string, getData func() (*bytes.Buffer, error)) error {

	if b, err := getData(); err == nil {

		result := r.db.Model(&Retry{}).Where("uuid = ?", UUID).Update("data", b.String())

		if int(result.RowsAffected) != 1 {
			var retry = Retry{UUID: UUID, Data: b.String()}

			return r.db.Create(&retry).Error
		}

	} else {
		return err
	}

	return nil
}

//ParseAll parse all element with callback parseData
func (r RStoreGorm) ParseAll(parseData func(string, []byte) error) {
	var retrys []Retry

	if err := r.db.Find(&retrys).Error; err == nil {
		for _, retry := range retrys {
			if err := parseData(retry.UUID, []byte(retry.Data)); err != nil {
				r.Delete(retry.UUID)
			}

		}

	}

}
