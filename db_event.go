package plugin

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

const (
	INSERT_EVENT = "insert"
	UPDATE_EVENT = "update"
)

type Config struct {
	DB          *gorm.DB
	AutoMigrate bool
	Tables      []string
}

// DB Event Plugin
type DBEvent struct {
	Config          *Config
	AuditableTables map[string]bool
}

func New(config Config) *DBEvent {
	dbEvent := DBEvent{
		Config:          &config,
		AuditableTables: map[string]bool{},
	}

	for _, table := range config.Tables {
		dbEvent.AuditableTables[table] = true
	}

	err := config.DB.AutoMigrate(
		&Version{},
	)

	if err != nil {
		panic(err)
	}

	return &dbEvent
}

// From `Plugin` interface
func (e *DBEvent) Name() string {
	return "gorm:db_event"
}

func (e *DBEvent) Initialize(db *gorm.DB) error {
	// Register a event for Create Callback from gorm.
	db.Callback().Create().Register("db_event:create", e.createCallback)
	db.Callback().Update().Register("db_event:update", e.updateCallback)
	return nil
}

// FIXME: maybe we could use db.Statement.Clauses["UPDATE/INSERT"] to judge the event name.
func (e *DBEvent) createCallback(db *gorm.DB) {
	// Check if the table needs to be tracked.
	if !e.AuditableTables[db.Statement.Schema.Name] {
		return
	}

	obj := getAuditableFields(db)

	// create a new version with serialized json.
	v, _ := json.Marshal(obj)
	version := Version{
		ItemID:   getCurrentItemID(db),
		ItemType: db.Statement.Schema.Name,
		Event:    INSERT_EVENT,
		Object:   v,
	}
	result := e.Config.DB.Create(&version)
	if result.Error != nil {
		fmt.Printf("Create Version Error: %s\n", result.Error)
	}
}

func (e *DBEvent) updateCallback(db *gorm.DB) {
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

	// create a new version with serialized json.
	objJSON, _ := json.Marshal(obj)
	prevObjJSON, _ := json.Marshal(prevObj)
	version := Version{
		ItemID:        itemID,
		ItemType:      db.Statement.Schema.Name,
		Event:         UPDATE_EVENT,
		Object:        objJSON,
		ObjectChanges: prevObjJSON,
	}
	e.Config.DB.Create(&version)
}

// Helper methods ===================

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
