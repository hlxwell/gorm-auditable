package main

import (
	"fmt"
	"net/http"

	auditable "github.com/hlxwell/gorm-auditable"
	"github.com/labstack/echo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Conn *gorm.DB

type User struct {
	gorm.Model
	Name string `gorm:"unique;auditable"`
}

func init() {
	setupConn()
}

func main() {
	e := echo.New()
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

func setupConn() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		"root", "", "localhost", "3306", "gorm_by_example",
	)

	var err error
	if Conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
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
