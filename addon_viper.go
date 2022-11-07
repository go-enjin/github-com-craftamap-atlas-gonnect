//go:build addon_viper || viper

package gonnect

// func NewAddon(configFile io.Reader, descriptorFile io.Reader) (a *Addon, err error) {
// 	var config *Profile
// 	var currentProfile string
// 	var addonDescriptor map[string]interface{}
//
// 	log.DebugF("Create new config object")
// 	if config, currentProfile, err = NewConfig(configFile); err != nil {
// 		log.ErrorF("Could not create new config object: %s\n", err)
// 		return
// 	}
//
// 	log.DebugF("Reading AddonDescriptor")
// 	if addonDescriptor, err = readAddonDescriptor(descriptorFile, config.BaseUrl); err != nil {
// 		log.ErrorF("Could not read AddonDescriptor: %s\n", err)
// 		return
// 	}
//
// 	log.DebugF("Creating new store")
// 	var s *store.Store
// 	if s, err = store.New(config.Store.Type, config.Store.DatabaseUrl); err != nil {
// 		log.ErrorF("Could not create new store: %s\n", err)
// 		return
// 	}
//
// 	return NewCustomAddon(config, currentProfile, addonDescriptor, s)
// }