package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	auditable "github.com/hlxwell/gorm-auditable"
	"github.com/labstack/echo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Conn *gorm.DB

type User struct {
	gorm.Model
	Name string `gorm:"unique;auditable"`
}

func main() {
	e := echo.New()
	makeConn("gorm_by_example")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("current_user_id", "12344321")
			return next(c)
		}
	})
	e.Use(auditable.GormInjector(Conn))
	e.GET("/hello", func(c echo.Context) error {
		conn := c.Get(auditable.GormDBKey).(*gorm.DB)
		conn.Create(&User{Name: "hello-hlxwell"})
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

// Helper Methods ============================

func makeConn(name string) {
	logLevel := logger.Error

	// custom logger
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logLevel,
			Colorful:      true,
		},
	)

	// data source name
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		"root",
		"",
		"localhost",
		"3306",
		name,
	)

	// Init conn
	var err error
	Conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   customLogger,
	})

	if err != nil {
		panic(err)
	}

	Conn.Use(auditable.New(auditable.Config{
		CurrentUserIDKey: "current_user_id",
		DB:               Conn,
		AutoMigrate:      false,
		Tables: []string{
			"User",
		},
	}))
}
