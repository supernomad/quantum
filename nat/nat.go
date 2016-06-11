package nat

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"math/big"
	"net"
)

type Nat struct {
	log       *logger.Logger
	privateIP []byte
	Mappings  map[uint64]common.Mapping
}

func (nat *Nat) ResolveOutgoing(payload *common.Payload) (*common.Payload, bool) {
	dip := big.NewInt(0)
	dip.SetBytes(payload.Packet[16:20])

	if mapping, ok := nat.Mappings[dip.Uint64()]; ok {
		payload.Mapping = mapping
		copy(payload.IpAddress, nat.privateIP)
		return payload, true
	}

	nat.log.Error("[NAT]", "Unknown Outgoing Mapping")
	return payload, false
}

func (nat *Nat) ResolveIncoming(payload *common.Payload) (*common.Payload, bool) {
	dip := big.NewInt(0)
	dip.SetBytes(payload.IpAddress)

	if mapping, ok := nat.Mappings[dip.Uint64()]; ok {
		payload.Mapping = mapping
		return payload, true
	}

	nat.log.Error("[NAT]", "Unknown Incoming Mapping")
	return payload, false
}

func New(privateIP string, mappings map[uint64]common.Mapping, log *logger.Logger) *Nat {
	return &Nat{
		log:       log,
		privateIP: net.ParseIP(privateIP).To4(),
		Mappings:  mappings,
	}
}
