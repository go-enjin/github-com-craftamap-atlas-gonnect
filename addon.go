package gonnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/github-com-craftamap-atlas-gonnect/store"
)

type Addon struct {
	Config          *Profile
	CurrentProfile  string
	Store           *store.Store
	AddonDescriptor map[string]interface{}
	Key             *string
	Name            *string
}

func readAddonDescriptor(descriptorReader io.Reader, baseUrl string) (map[string]interface{}, error) {
	vals := map[string]string{
		"BaseUrl": baseUrl,
	}

	temp, err := ioutil.ReadAll(descriptorReader)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("descriptor").Parse(string(temp))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer

	err = tmpl.ExecuteTemplate(&buffer, "descriptor", vals)
	if err != nil {
		return nil, err
	}

	descriptor := map[string]interface{}{}

	err = json.Unmarshal(buffer.Bytes(), &descriptor)
	if err != nil {
		return nil, err
	}

	return descriptor, nil
}

func NewCustomAddon(config *Profile, currentProfile string, addonDescriptor map[string]interface{}, s *store.Store) (a *Addon, err error) {
	log.InfoF("Initializing new Addon with profile: %v", currentProfile)
	log.DebugF("Using Addon Profile: %v", config)
	log.DebugF("Using Addon descriptor: %v", addonDescriptor)

	var ok bool
	var name, key string
	if name, ok = addonDescriptor["name"].(string); !ok {
		err = fmt.Errorf("name could not be read from AddonDescriptor")
		return
	}

	if key, ok = addonDescriptor["key"].(string); !ok {
		err = fmt.Errorf("key could not be read from AddonDescriptor")
		return
	}

	a = &Addon{
		Config:          config,
		Store:           s,
		CurrentProfile:  currentProfile,
		AddonDescriptor: addonDescriptor,
		Name:            &name,
		Key:             &key,
	}

	log.DebugF("addon successfully initialized")
	return
}
