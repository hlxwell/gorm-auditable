# GORM Auditable

The purpose of Gorm-Auditable project for solving users wants to know the difference between each change, also who made those changes. This project is inspired by [Paper Trail](https://github.com/paper-trail-gem/paper_trail). The difference from [QOR audited](https://github.com/qor/audited) is QOR audited only records who made the last change but cannot show the difference.

## Features

- Be able to record each new version of your tracked database records. (only support insert and update for now)
- Be able to track who did the change.

## How to use

### 1. Config the Auditable

Add GORM plugin with config:

```go
db.Use(auditable.New(auditable.Config{
  CurrentUserIDKey: "current_user_id",  // Current User ID Key from echo.Context, which is for plugin to get current operator id.
  DB:               Conn,               // Database Connection
  AutoMigrate:      true,               // Do you need *versions* table to be created automatically?
  Tables: []string{                     // All the tables you would like to track versions.
    "User",
  },
}))
```

### 2. Config field you would like to track.
In your gorm model, you need to add the *auditable* to the field, otherwise, it won't be recorded:

```go
type User struct {
  gorm.Model
  Name string `gorm:"unique;auditable"`
}
```

### 3. Add echo middleware

If you are using echo framework, there already is a middleware that could inject scoped gorm object into `Context`, which is used to set `current_user_id` when inserting the `Version` record.

```go
e.Use(auditable.GormInjector(Conn))
```

Please remember, you need to add your authenticate middleware before it.

```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
  return func(c echo.Context) error {
    // e.g. You need to set the current_user_id before Gorm Injector.
    c.Set("current_user_id", "12344321")
    return next(c)
  }
})

e.Use(auditable.GormInjector(Conn))
```

### 4. Use gorm connection from echo context

```go
conn := c.Get(auditable.GormDBKey).(*gorm.DB)
conn.Create(&User{Name: "Michael He"})
```

## Example Code

You could run `go run example/example.go` to try the example. But make sure to change the database config.