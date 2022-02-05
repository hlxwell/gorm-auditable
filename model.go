package auditable

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Version struct {
	gorm.Model
	ItemType      string
	ItemID        uint
	Event         string
	Whodunnit     string
	Object        datatypes.JSON
	ObjectChanges datatypes.JSON
}
