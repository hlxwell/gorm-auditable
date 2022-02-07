package main

import (
	"fmt"
	"net/http"
	"time"

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
			c.Set("current_user_id", fmt.Sprintf("%d", time.Now().Nanosecond()))
			return next(c)
		}
	})
	e.Use(auditable.GormInjector(Conn))
	e.GET("/insert", func(c echo.Context) error {
		conn := c.Get(auditable.GormDBKey).(*gorm.DB)
		conn.Create(&User{Name: "hello-hlxwell"})
		return c.String(http.StatusOK, "User Created!")
	})
	e.GET("/update", func(c echo.Context) error {
		conn := c.Get(auditable.GormDBKey).(*gorm.DB)
		var user User
		conn.First(&user)
		user.Name = fmt.Sprintf("name-%d", time.Now().Nanosecond())
		conn.Save(&user)
		return c.String(http.StatusOK, "Update Successfully!")
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

	if err = Conn.AutoMigrate(&User{}); err != nil {
		panic(err)
	}

	Conn.Use(auditable.New(auditable.Config{
		CurrentUserIDKey: "current_user_id",
		DB:               Conn,
		AutoMigrate:      true,
		Tables: []string{
			"User",
		},
	}))
}
