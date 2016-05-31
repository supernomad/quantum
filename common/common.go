package common

import "encoding/json"

type Payload struct {
	Packet    []byte
	PublicKey string
	Address   string
}

type Mapping struct {
	Address   string
	PublicKey string
}

func (m Mapping) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

func ParseMapping(data string) Mapping {
	var mapping Mapping
	json.Unmarshal([]byte(data), &mapping)
	return mapping
}
