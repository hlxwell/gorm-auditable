package auditable

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

const (
	InsertEvent = "insert"
	UpdateEvent = "update"
	UserIDKey   = "auditable:current_user"
	GormDBKey   = "GORM_DB"
)

// Keep the config as a global variable.
var config Config

type Config struct {
	CurrentUserIDKey string
	DB               *gorm.DB
	AutoMigrate      bool
	Tables           []string
}

// DB Event Plugin
type DBEvent struct {
	Config          *Config
	AuditableTables map[string]bool
}

func New(cfg Config) *DBEvent {
	config = cfg
	dbEvent := DBEvent{
		Config:          &config,
		AuditableTables: map[string]bool{},
	}

	for _, table := range config.Tables {
		dbEvent.AuditableTables[table] = true
	}

	if config.AutoMigrate {
		err := config.DB.AutoMigrate(
			&Version{},
		)

		if err != nil {
			panic(err)
		}
	}

	return &dbEvent
}

// From `Plugin` interface
func (e *DBEvent) Name() string {
	return "gorm:db_event"
}

// Register callback event for After Create & Update Callback in gorm.
func (e *DBEvent) Initialize(db *gorm.DB) error {
	db.Callback().Create().After("gorm:create").Register("db_event:create", e.createCallback)
	db.Callback().Update().After("gorm:update").Register("db_event:update", e.updateCallback)
	return nil
}

// FIXME: maybe we could use db.Statement.Clauses["UPDATE/INSERT"] to judge the event name.
func (e *DBEvent) createCallback(db *gorm.DB) {
	// If creation failed, then return.
	if db.RowsAffected == 0 {
		return
	}

	// Check if the table needs to be tracked.
	if !e.AuditableTables[db.Statement.Schema.Name] {
		return
	}

	obj := getAuditableFields(db)
	// get current operator id
	userID, _ := db.Get(UserIDKey)

	// create a new version with serialized json.
	v, _ := json.Marshal(obj)
	version := Version{
		ItemID:    getCurrentItemID(db),
		ItemType:  db.Statement.Schema.Name,
		Event:     InsertEvent,
		Object:    v,
		Whodunnit: fmt.Sprintf("%v", userID),
	}

	result := e.Config.DB.Create(&version)
	if result.Error != nil {
		fmt.Printf("Create Version Error: %s\n", result.Error)
	}
}

func (e *DBEvent) updateCallback(db *gorm.DB) {
	// If updating failed, then return.
	if db.RowsAffected == 0 {
		return
	}

	// Ignore invalid updates
	itemID := getCurrentItemID(db)
	if itemID == 0 {
		return
	}

	// Check if the table needs to be tracked.
	if !e.AuditableTables[db.Statement.Schema.Name] {
		return
	}

	// Add all auditable fields to the final serialized object.
	obj := getAuditableFields(db)

	// Find the previous version auditable fields values.
	prevObj := map[string]interface{}{}
	var prevVersion Version
	result := e.Config.DB.Where("item_type = ? AND item_id = ?", db.Statement.Schema.Name, itemID).Last(&prevVersion)

	// If can find the record, then fill the prevObj for all the auditable fields for it.
	if result.Error == nil {
		var prevData map[string]interface{}
		json.Unmarshal([]byte(prevVersion.Object.String()), &prevData)

		for _, field := range db.Statement.Schema.Fields {
			// Check if it is the auditable field.
			if len(field.TagSettings["AUDITABLE"]) == 0 {
				continue
			}

			prevObj[field.DBName] = prevData[field.DBName]
		}
	} else {
		fmt.Println(result.Error)
	}

	// get current operator id
	userID, _ := db.Get(UserIDKey)

	// create a new version with serialized json.
	objJSON, _ := json.Marshal(obj)
	changesJSON, _ := getFieldChanges(prevObj, obj)
	version := Version{
		ItemID:        itemID,
		ItemType:      db.Statement.Schema.Name,
		Event:         UpdateEvent,
		Object:        objJSON,
		ObjectChanges: changesJSON,
		Whodunnit:     fmt.Sprintf("%v", userID),
	}
	e.Config.DB.Create(&version)
}

// Helper methods ===================

// Get Item ID in current DB statement
func getCurrentItemID(db *gorm.DB) uint {
	itemID, isEmpty := db.Statement.Schema.FieldsByDBName["id"].ValueOf(db.Statement.ReflectValue)
	if isEmpty {
		return 0
	}

	result, ok := itemID.(uint)
	if ok {
		return result
	}
	return 0
}

// getAuditableFields will extract all auditable fields to a map hash.
func getAuditableFields(db *gorm.DB) map[string]interface{} {
	obj := map[string]interface{}{}
	for _, field := range db.Statement.Schema.Fields {
		// Check if it is the auditable field.
		if len(field.TagSettings["AUDITABLE"]) == 0 {
			continue
		}

		if fieldValue, isZero := field.ValueOf(db.Statement.ReflectValue); !isZero {
			obj[field.DBName] = fieldValue
		}
	}

	return obj
}

// Generate a map to list all the field changes.
func getFieldChanges(prevItem map[string]interface{}, currentItem map[string]interface{}) ([]byte, error) {
	result := make(map[string][]interface{})

	for k, v := range prevItem {
		result[k] = []interface{}{v, currentItem[k]}
	}

	return json.Marshal(result)
}
