package backend

import (
	"errors"
	"github.com/Supernomad/quantum/common"
	"sync"
)

const (
	// LIBKV backend type
	LIBKV int = 0
)

// Backend interface for extendability
type Backend interface {
	GetMapping(ip uint32) (*common.Mapping, bool)
	Init() error
	Start(wg *sync.WaitGroup)
	Stop()
}

// New Bakend object
func New(kind int, log *common.Logger, cfg *common.Config) (Backend, error) {
	switch kind {
	case LIBKV:
		return newLibkv(log, cfg)
	default:
		return nil, errors.New("non compatible backend specified")
	}
}
