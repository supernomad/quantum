package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"github.com/supernomad/quantum/common"
)

// Sections is cool
type Sections struct {
	Sections []*Section `json:"sections"`
}

func (s *Sections) get(name string) (*Section, bool) {
	for i := 0; i < len(s.Sections); i++ {
		if s.Sections[i].Name == name {
			return s.Sections[i], true
		}
	}
	return nil, false
}

// Section is cool
type Section struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Options     []*Option `json:"options"`
}

// Option is cool
type Option struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Short          string `json:"short"`
	Long           string `json:"long"`
	Default        string `json:"default"`
	Type           string `json:"type"`
	TypeDefinition string `json:"type_def"`
}

var sectionDescriptions = map[string]string{
	"General":   "The General configuration section focuses on the base level configuration options to modify the bahvior of ``quantum`` at a fundamental level.",
	"Plugins":   "The Plugins configuration section provides options to configure the internal ``quantum`` plugins. ",
	"Datastore": "The Datastore configuration section modifies how the backend datastore is interacted with.",
	"DTLS":      "The DTLS configuration section only applies when the networking backend for ``quantum`` is configured to use 'dtls'. This section configures the backend so that it can properly communicate with the other peers in the network.",
	"Stats":     "The Stats section exposes options to change how the REST API that ``quantum`` runs internally is exported.",
	"Network":   "The Network configuration allows setting up the defaults for the entire ``quantum`` network.",
}

var typeDefinitions = map[string]string{
	"string":   "A basic string type, strings with spaces or special characters should be quoted.",
	"int":      "A basic integer type, accepts any integer value.",
	"list":     "A basic list type, accepts a comma delimited list of values.",
	"ip":       "A basic ip type, which accepts both IPv4 and IPv6 addresses where specified.",
	"ip-list":  "A special list type specific to ip type options.",
	"duration": "A duration value, syntax examples being '1s', '2h', '3d'.",
	"bool":     "A flag that takes no value.",
}

func parseField(tag reflect.StructTag) (internal, fieldType, short, long, def, description, section, name string) {
	internal = tag.Get("internal")
	fieldType = tag.Get("type")
	short = tag.Get("short")
	long = tag.Get("long")
	def = tag.Get("default")
	description = tag.Get("description")
	section = tag.Get("section")
	name = tag.Get("name")
	return
}

func main() {
	cfg := &common.Config{}

	sections := &Sections{
		Sections: []*Section{},
	}

	st := reflect.TypeOf(*cfg)
	sv := reflect.ValueOf(cfg).Elem()

	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		fieldValue := sv.Field(i)
		internal, fieldType, short, long, def, description, sectionName, name := parseField(field.Tag)

		if internal == "true" || !fieldValue.CanSet() {
			continue
		}

		option := &Option{
			Name:           name,
			Description:    description,
			Short:          short,
			Long:           long,
			Default:        def,
			Type:           fieldType,
			TypeDefinition: typeDefinitions[fieldType],
		}
		if section, ok := sections.get(sectionName); !ok {
			section := &Section{
				Name:        sectionName,
				Description: sectionDescriptions[sectionName],
				Options:     []*Option{option},
			}
			sections.Sections = append(sections.Sections, section)
		} else {
			section.Options = append(section.Options, option)
		}
	}

	data, err := json.MarshalIndent(sections, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Writing json file to: %s", os.Args[1])
	ioutil.WriteFile(os.Args[1], data, 0644)
}
