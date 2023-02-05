package util

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

type MarshallType uint

const (
	Json MarshallType = iota
	Yaml
)

func Marshall(typ MarshallType, in any) ([]byte, error) {
	var outb []byte
	var err error
	switch typ {
	case Json:
		outb, err = json.Marshal(in)
	case Yaml:
		outb, err = yaml.Marshal(in)
	}
	return outb, err
}

func Unmarshall(typ MarshallType, inb []byte, out any) error {
	var err error
	switch typ {
	case Json:
		err = json.Unmarshal(inb, out)
	case Yaml:
		err = yaml.Unmarshal(inb, out)
	}
	return err
}
