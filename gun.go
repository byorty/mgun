package mgun

import (
	"time"
	"regexp"
)

var (
	arrayParamRegexp = regexp.MustCompile(`[\w\d\-\_]\[\]+`)
	configParamRegexp = regexp.MustCompile(`\$\{([\w\d\-\_\.]+)\}`)
)

type Gun struct {
	Concurrency  int           `yaml:"concurrency"`
	LoopCount    int           `yaml:"loopCount"`
	Timeout      time.Duration `yaml:"timeout"`
	Scheme       string        `yaml:"scheme"`
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Features     Features      `yaml:"headers"`
	Calibers     CaliberMap    `yaml:"params"`
	Cartridges   Cartridges    `yaml:"requests"`
}

func NewGun() *Gun {
	gun := new(Gun)
	gun.Features = make(Features, 0)
	gun.Calibers = make(CaliberMap)
	gun.Cartridges = make(Cartridges, 0)
	return gun
}

type CaliberMap map[string]*Caliber
type CaliberList []CaliberMap

func (this CaliberMap) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	calibers := make(map[string]interface{})
	err := unmarshal(calibers)

	if rawSessions, ok := calibers["session"].([]interface{}); ok {
		delete(calibers, "session")
		list := make(CaliberList, 0)
		for _, rawSession := range rawSessions {
			if session, ok := rawSession.(map[string]interface{}); ok {
				caliberMap := make(CaliberMap)
				caliberMap.fill(session)
				list = append(list, caliberMap)
			}
		}
		caliber := new(Caliber)
		caliber.kind = CALIBER_KIND_SESSION
		caliber.children = list
		this["session"] = caliber
	}

	this.fill(calibers)
	return err
}

func (this CaliberMap) fill(rawCalibers map[string]interface{}) {
	for key, value := range rawCalibers {
		caliber := new(Caliber)
		_, isArray := value.([]interface{})
		if arrayParamRegexp.MatchString(key) && isArray {
			caliber.kind = CALIBER_KIND_MULTIPLE
		} else {
			caliber.kind = CALIBER_KIND_SIMPLE
		}
		caliber.size = value
		this[key] = caliber
	}
}

type Caliber struct {
	kind     CaliberKind
	size     interface{}
	children CaliberList
}

type CaliberKind int

const (
	CALIBER_KIND_SIMPLE CaliberKind = iota
	CALIBER_KIND_MULTIPLE
	CALIBER_KIND_SESSION
)

type Cartridges []*Cartridge

func (this *Cartridges) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	rawCartridges := make([]map[string]interface{}, 0)
	err := unmarshal(&rawCartridges)

	for _, rawCartridge := range rawCartridges {
		cartridge := new(Cartridge)
		for key, value := range rawCartridge {
			switch (key) {
			case GET_METHOD, POST_METHOD, PUT_METHOD, DELETE_METHOD:
				cartridge.path = NewDescribedFeature(key, value.(string))
				break;
			case RANDOM_METHOD, SYNC_METHOD:
				cartridge.path = NewFeature(key)
				break;
			case "headers":
				cartridge.bulletFeatures = make(Features, 0)
				cartridge.bulletFeatures.fill(value.(map[interface{}]interface{}))
				break;
			case "params":
				cartridge.chargeFeatures = make(Features, 0)
				cartridge.chargeFeatures.fill(value.(map[interface{}]interface{}))
				break;
			case "timeout":
				cartridge.timeout = time.Duration(value.(int))
				break;
			}
		}
		*this = append(*this, cartridge)
	}

	return err
}

const (
	GET_METHOD     = "GET"
	POST_METHOD    = "POST"
	PUT_METHOD     = "PUT"
	DELETE_METHOD  = "DELETE"
	RANDOM_METHOD  = "RANDOM"
 	SYNC_METHOD    = "SYNC"
	INCLUDE_METHOD = "INCLUDE"
)

type Cartridge struct {
	path           *Feature
	bulletFeatures Features
	chargeFeatures Features
	timeout        time.Duration
	children       Cartridges
}

type FeatureKind int

const (
	FEATURE_KIND_SIMPLE FeatureKind = iota
	FEATURE_KIND_MULTIPLE
)

type Features []*Feature

func (this *Features) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	rawFeatures := make(map[interface{}]interface{})
	err := unmarshal(&rawFeatures)

	this.fill(rawFeatures)

	return err
}

func (this *Features) fill(rawFeatures map[interface{}]interface{}) {
	for rawKey, rawValue := range rawFeatures {
		key := rawKey.(string)
		value := rawValue.(string)
		feature := NewFeature(key)
		if configParamRegexp.MatchString(value) {
			feature.description = configParamRegexp.ReplaceAllString(value, "%v")
			feature.units = configParamRegexp.FindAllString(value, -1)
			feature.kind = FEATURE_KIND_MULTIPLE
		} else {
			feature.description = value
			feature.kind = FEATURE_KIND_SIMPLE
		}
		*this = append(*this, feature)
	}
}

type Feature struct {
	name        string
	description string
	units       []string
	kind        FeatureKind
}

func NewFeature(name string) *Feature {
	feature := new(Feature)
	feature.name = name
	return feature
}

func NewDescribedFeature(name, description string) *Feature {
	feature := NewFeature(name)
	feature.description = description
	return feature
}




