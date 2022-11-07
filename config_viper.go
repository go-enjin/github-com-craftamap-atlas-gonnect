//go:build config_viper || viper

package gonnect

// import (
// 	"io"
//
// 	"github.com/spf13/viper"
//
// 	"github.com/go-enjin/be/pkg/log"
// )
//
// func NewConfig(configFile io.Reader) (*Profile, string, error) {
// 	log.DebugF("Initializing Configuration")
//
// 	runtimeViper := viper.New()
// 	runtimeViper.SetDefault("CurrentProfile", "dev")
// 	_ = runtimeViper.BindEnv("CurrentProfile", "GONNECT_PROFILE")
// 	runtimeViper.SetConfigType("json")
// 	config := &Config{}
//
// 	err := runtimeViper.ReadConfig(configFile)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	err = runtimeViper.Unmarshal(config)
// 	if err != nil {
// 		return nil, "", err
// 	}
//
// 	if config.CurrentProfile == "" {
// 		return nil, "", ErrConfigNoProfileSelected
// 	}
//
// 	if profile, ok := config.Profiles[config.CurrentProfile]; !ok {
// 		return nil, "", ErrConfigProfileNotFound
// 	} else {
// 		return &profile, config.CurrentProfile, nil
// 	}
// }