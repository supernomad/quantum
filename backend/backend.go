package backend

import (
	"errors"
	"github.com/Supernomad/quantum/common"
)

const (
	// LIBKV backend type
	LIBKV int = 0
)

// Backend interface for extendability
type Backend interface {
	GetMapping(ip uint32) (*common.Mapping, bool)
	Init() error
	Start()
	Stop()
}

// New Bakend object
func New(kind int, cfg *common.Config) (Backend, error) {
	switch kind {
	case LIBKV:
		return newLibkv(cfg)
	default:
		return nil, errors.New("Non compatible backend specified.")
	}
}
