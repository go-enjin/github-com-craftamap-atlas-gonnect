package gonnect

import (
	"errors"
)

var ErrConfigNoProfileSelected = errors.New("No Profile selected; Set CurrentProfile in the config file or set GONNECT_PROFILE")
var ErrConfigProfileNotFound = errors.New("Profile not found!")

type Config struct {
	CurrentProfile string
	Profiles       map[string]Profile
}

type Profile struct {
	BaseUrl       string
	Store         StoreConfiguration
	SignedInstall bool
}

func NewProfile(baseUrl, dbType, dbUri string, signedInstall bool) *Profile {
	return &Profile{
		BaseUrl: baseUrl,
		Store: StoreConfiguration{
			Type:        dbType,
			DatabaseUrl: dbUri,
		},
		SignedInstall: signedInstall,
	}
}

type StoreConfiguration struct {
	Type        string
	DatabaseUrl string
}

func NewConfiguration(dbType, dbUrl string) StoreConfiguration {
	return StoreConfiguration{
		Type:        dbType,
		DatabaseUrl: dbUrl,
	}
}
