package nat

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"math/big"
)

type Nat struct {
	log      *logger.Logger
	Mappings map[uint64]common.Mapping
}

func (nat *Nat) ResolveOutgoing(payload *common.Payload) (*common.Payload, bool) {
	if payload.Packet[0]>>4 != 4 {
		nat.log.Error("[NAT]", "Unknown IP version recieved")
		return payload, false
	}

	dip := big.NewInt(0)
	dip.SetBytes(payload.Packet[16:20])

	if mapping, ok := nat.Mappings[dip.Uint64()]; ok {
		payload.Address = mapping.Address
		payload.PublicKey = mapping.PublicKey
		return payload, true
	}
	return payload, false
}

func (nat *Nat) ResolveIncoming(payload *common.Payload) (*common.Payload, bool) {
	return payload, true
}

func New(mappings map[uint64]common.Mapping, log *logger.Logger) *Nat {
	return &Nat{
		log:      log,
		Mappings: mappings,
	}
}
