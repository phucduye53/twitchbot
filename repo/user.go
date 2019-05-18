package repo

import (
	db "twitchbot/db"
	"twitchbot/models"

	"github.com/jinzhu/gorm"
)

type IUser interface {
	Create(*models.User) error
	Update(*models.User) error
	Delete(*models.User) error
	GetById(int) (*models.User, error)
}

func Create(user *models.User) error {

	err := db.GetDBcontext().Model(&models.User{}).Create(user).Error
	if err != nil {
		db.GetDBcontext().Rollback()
	}
	return err
}
func Update(user *models.User) error {
	err := db.GetDBcontext().Model(&models.User{}).Update(user).Error
	if err != nil {
		db.GetDBcontext().Rollback()
	}
	return err
}
func Delete(user *models.User) error {
	err := db.GetDBcontext().Model(&models.User{}).Delete(user).Error
	if err != nil {
		db.GetDBcontext().Rollback()
	}
	return err
}
func GetByName(strName string) (*models.User, error) {
	var user models.User
	err := db.GetDBcontext().Model(&models.User{}).Where("name = ?", strName).Limit(1).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &user, err

}
