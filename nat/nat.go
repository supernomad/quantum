package nat

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"net"
)

type Nat struct {
	log      *logger.Logger
	Mappings map[string]common.Mapping
}

func (nat *Nat) ResolveOutgoing(payload common.Payload) common.Payload {
	var dip string
	switch payload.Packet[0] >> 4 {
	case 4:
		dip = net.IP(payload.Packet[16:20]).String()
	case 6:
		dip = net.IP(payload.Packet[24:40]).String()
	default:
		nat.log.Error("[NAT] Unknown IP version recieved")
		return payload
	}

	if mapping, ok := nat.Mappings[dip]; ok {
		payload.Address = mapping.Address
		payload.PublicKey = mapping.PublicKey
	}
	return payload
}

func (nat *Nat) ResolveIncoming(payload common.Payload) common.Payload {
	return payload
}

func New(mappings map[string]common.Mapping, log *logger.Logger) *Nat {
	return &Nat{
		log:      log,
		Mappings: mappings,
	}
}
