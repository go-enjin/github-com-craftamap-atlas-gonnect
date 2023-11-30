package store

import (
	"github.com/go-enjin/be/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DefaultTableName = "atlas_gonnect_tenants"

type Store struct {
	Database *gorm.DB
	table    string
}

func New(dbType string, databaseUrl string) (store *Store, err error) {
	log.TraceF("Initializing Database Connection")
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
	store, err = NewTableFrom(DefaultTableName, db)
	return
}

func NewTableFrom(table string, db *gorm.DB) (store *Store, err error) {
	store = &Store{
		table:    table,
		Database: db,
	}
	log.TraceF("Migrating Database Schemas")
	if err = store.Tx().AutoMigrate(&Tenant{}); err != nil {
		return
	}
	log.TraceF("Database Connection initialized")
	return
}

func NewMustTableFrom(table string, db *gorm.DB) (store *Store) {
	var err error
	if store, err = NewTableFrom(table, db); err != nil {
		log.FatalDF(1, "%v", err)
		return
	}
	return
}

func (s *Store) Tx() (tx *gorm.DB) {
	tx = s.Database.Scopes(func(tx *gorm.DB) *gorm.DB {
		if s.table == "" {
			return tx.Table(DefaultTableName)
		}
		return tx.Table(s.table)
	})
	return
}

func (s *Store) Get(clientKey string) (*Tenant, error) {
	tenant := Tenant{}
	log.TraceF("Tenant with clientKey %s requested from database", clientKey)
	if result := s.Tx().Where(&Tenant{ClientKey: clientKey}).First(&tenant); result.Error != nil {
		return nil, result.Error
	}
	log.TraceF("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) GetByUrl(url string) (*Tenant, error) {
	tenant := Tenant{}
	log.TraceF("Tenant with clientKey %s requested from database", url)
	if result := s.Tx().Where(&Tenant{BaseURL: url}).First(&tenant); result.Error != nil {
		return nil, result.Error
	}
	log.TraceF("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) Set(tenant *Tenant) (*Tenant, error) {
	log.DebugF("Tenant %+v will be inserted or updated in database", tenant)

	optionalExistingRecord := Tenant{}
	if result := s.Tx().Where(&Tenant{ClientKey: tenant.ClientKey}).First(&optionalExistingRecord); result.Error != nil {
		// If no entry matching the clientKey exists, insert the tenant,
		// otherwise update the tenant
		log.DebugF("Tenant %+v will be inserted in database", tenant)
		if result := s.Tx().Create(tenant); result.Error != nil {
			return nil, result.Error
		}
	} else {
		log.DebugF("Tenant %+v will be updated in database", tenant)
		if result := s.Tx().Model(tenant).Where(&Tenant{ClientKey: tenant.ClientKey}).Updates(tenant).Update("AddonInstalled", tenant.AddonInstalled); result.Error != nil {
			return nil, result.Error
		}
	}

	log.TraceF("Tenant %+v successfully inserted or updated", tenant)
	return tenant, nil
}

func (s *Store) Delete(clientKey string) (err error) {
	tenant := Tenant{}
	if result := s.Tx().Where(&Tenant{ClientKey: clientKey}).First(&tenant); result.Error != nil {
		return result.Error
	}
	log.WarnF("deleting tenant with clientKey %s from database", clientKey)
	return s.Tx().Delete(&tenant).Error
}