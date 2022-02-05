package auditable

import (
	"github.com/labstack/echo"
	"gorm.io/gorm"
)

func GormInjector(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// default will use the empty scope.
			c.Set(GormDBKey, db)
			if userID := c.Get(config.CurrentUserIDKey); userID != nil {
				// Set Scoped Gorm object into echo context.
				c.Set(GormDBKey, db.Set(UserIDKey, userID))
			}
			return next(c)
		}
	}
}
