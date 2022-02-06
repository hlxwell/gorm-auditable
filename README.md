# GORM Auditable

The purpose of Gorm-Auditable project for solving user want to know the difference between each changes, also who made those changes. This project is inspired by [Paper Trail](https://github.com/paper-trail-gem/paper_trail). The difference from [QOR audited](https://github.com/qor/audited) is QOR audited only record who made the last change, but cannot show the difference.

## How to use

### 1. Config the Auditable

Add GORM plugin with config:

```
db.Use(auditable.New(auditable.Config{
  CurrentUserIDKey: "current_user_id",  // Current User ID Key from echo.Context, which is for plugin to get current operator id.
  DB:               Conn,               // Database Connection
  AutoMigrate:      true,               // Do you need Versions table to be created automatically?
  Tables: []string{                     // All the tables you would like to track versions.
    "User",
  },
}))
```

### 2. Config field you would like to track.
In your gorm model, you need to add the *auditable* to the field, otherwise it won't be recorded:

```
type User struct {
	gorm.Model
	Name string `gorm:"unique;auditable"`
}
```

### 3. Add echo middleware

If you are using echo framework, there already is a middleware could inject scoped gorm object into `Context`, which is used to set `current_user_id` when insert the `Version` record.

```
e.Use(auditable.GormInjector(Conn))
```

Please remember, you need to add your authenticate middleware before it.

```
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
  return func(c echo.Context) error {
    // e.g. You need to set the current_user_id before Gorm Injector.
    c.Set("current_user_id", "12344321")
    return next(c)
  }
})

e.Use(auditable.GormInjector(Conn))
```

## Example Code

You could run `go run example/example.go` to try the example. But make sure to change the database config.