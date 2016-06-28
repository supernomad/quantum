package config

import (
	"flag"
)

type Config struct {
	PrivateIP     string
	PublicIP      string
	SubnetMask    string
	EtcdHost      string
	EtcdKey       string
	InterfaceName string
	ListenAddress string
	ListenPort    int
	EnableCrypto  bool
}

func New() *Config {
	laddr := flag.String("listen-address", "0.0.0.0", "The ip address to listen on for forwarded packets.")
	lport := flag.Int("listen-port", 1099, "The ip port to listen on for forwarded packets.")

	ifaceName := flag.String("interface-name", "quantum", "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")

	etcdHost := flag.String("etcd-host", "http://127.0.0.1:2379", "The etcd endpoint to use as a configuration hub.")
	etcdKey := flag.String("etcd-key", "/quantum", "The etcd key that quantum information is stored under.")

	crypto := flag.Bool("crypto", true, "Whether or not to encrypt data sent and recieved, by this node, to and from the rest of the cluster.")

	subnetMask := flag.String("subnet-mask", "16", "The subnet mask in either ip format or bit width format")
	privateIP := flag.String("private-ip", "", "The private ip address of this node.")
	publicIP := flag.String("public-ip", "", "The public ip address of this node.")

	flag.Parse()

	return &Config{
		PrivateIP:     *privateIP,
		PublicIP:      *publicIP,
		SubnetMask:    *subnetMask,
		EtcdHost:      *etcdHost,
		EtcdKey:       *etcdKey,
		InterfaceName: *ifaceName,
		ListenAddress: *laddr,
		ListenPort:    *lport,
		EnableCrypto:  *crypto,
	}
}
