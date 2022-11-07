package store

import (
	"github.com/go-enjin/be/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	Database *gorm.DB
}

func New(dbType string, databaseUrl string) (store *Store, err error) {
	log.DebugF("Initializing Database Connection")
	var dialect gorm.Dialector
	switch dbType {
	case "postgres":
		dialect = postgres.Open(databaseUrl)
	case "mysql":
		dialect = mysql.Open(databaseUrl)
	default:
		dialect = sqlite.Open(databaseUrl)
	}

	var db *gorm.DB
	if db, err = gorm.Open(dialect); err != nil {
		return
	}

	store, err = NewFrom(db)
	return
}

func NewFrom(db *gorm.DB) (store *Store, err error) {
	log.DebugF("Migrating Database Schemas")
	if err = db.AutoMigrate(&Tenant{}); err != nil {
		return
	}
	store = &Store{
		Database: db,
	}
	log.DebugF("Database Connection initialized")
	return
}

func (s *Store) Get(clientKey string) (*Tenant, error) {
	tenant := Tenant{}
	log.DebugF("Tenant with clientKey %s requested from database", clientKey)
	if result := s.Database.Where(&Tenant{ClientKey: clientKey}).First(&tenant); result.Error != nil {
		return nil, result.Error
	}
	log.DebugF("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) GetByUrl(url string) (*Tenant, error) {
	tenant := Tenant{}
	log.DebugF("Tenant with clientKey %s requested from database", url)
	if result := s.Database.Where(&Tenant{BaseURL: url}).First(&tenant); result.Error != nil {
		return nil, result.Error
	}
	log.DebugF("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) Set(tenant *Tenant) (*Tenant, error) {
	log.DebugF("Tenant %+v will be inserted or updated in database", tenant)

	optionalExistingRecord := Tenant{}
	if result := s.Database.Where(&Tenant{ClientKey: tenant.ClientKey}).First(&optionalExistingRecord); result.Error != nil {
		// If no entry matching the clientKey exists, insert the tenant,
		// otherwise update the tenant
		log.DebugF("Tenant %+v will be inserted in database", tenant)
		if result := s.Database.Create(tenant); result.Error != nil {
			return nil, result.Error
		}
	} else {
		log.DebugF("Tenant %+v will be updated in database", tenant)
		if result := s.Database.Model(tenant).Where(&Tenant{ClientKey: tenant.ClientKey}).Updates(tenant).Update("AddonInstalled", tenant.AddonInstalled); result.Error != nil {
			return nil, result.Error
		}
	}

	log.DebugF("Tenant %+v successfully inserted or updated", tenant)
	return tenant, nil
}

func (s *Store) Delete(clientKey string) (err error) {
	tenant := Tenant{}
	if result := s.Database.Where(&Tenant{ClientKey: clientKey}).First(&tenant); result.Error != nil {
		return result.Error
	}
	log.DebugF("deleting tenant with clientKey %s from database", clientKey)
	return s.Database.Delete(&tenant).Error
}
