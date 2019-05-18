package db

import (
	"fmt"
	"twitchbot/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

var dbContext *gorm.DB

func ConnectToDb() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	dbContext, err := gorm.Open("postgres", psqlInfo)

	if err != nil {

	}

	dbContext.AutoMigrate(&models.User{})

	return dbContext

}

func GetDBcontext() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	dbContext, err := gorm.Open("postgres", psqlInfo)

	if err != nil {

	}

	dbContext.AutoMigrate(&models.User{})

	return dbContext
}
