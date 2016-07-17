package config

import (
	"flag"
	"strings"
	"time"
)

type Config struct {
	InterfaceName string
	PrivateIP     string
	PublicIP      string
	SubnetMask    string

	ListenAddress string
	ListenPort    int

	Prefix       string
	LeaseTime    time.Duration
	SyncInterval time.Duration
	Retries      time.Duration
	EnableCrypto bool

	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string
}

func New() *Config {
	ifaceName := flag.String("interface-name", "quantum", "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")
	privateIP := flag.String("private-ip", "", "The private ip address of this node.")
	publicIP := flag.String("public-ip", "", "The public ip address of this node.")
	subnetMask := flag.String("subnet-mask", "16", "The subnet mask in bit width format")

	laddr := flag.String("listen-address", "0.0.0.0", "The ip address to listen on for forwarded packets.")
	lport := flag.Int("listen-port", 1099, "The ip port to listen on for forwarded packets.")

	prefix := flag.String("prefix", "/quantum", "The etcd key that quantum information is stored under.")
	leaseTime := flag.Duration("lease-time", 300, "Lease time for the private ip address.")
	syncInterval := flag.Duration("sync-interval", 30, "The backend sync interval")
	retries := flag.Duration("retries", 5, "The number of times to retry aquiring the private ip address lease.")
	crypto := flag.Bool("crypto", true, "Whether or not to encrypt data sent and recieved, by this node, to and from the rest of the cluster.")

	etcdEndpoints := flag.String("etcd-endpoints", "http://127.0.0.1:2379", "The etcd endpoints to use, in a comma separated list.")
	etcdUsername := flag.String("etcd-username", "", "The etcd user to use for authentication.")
	etcdPassword := flag.String("etcd-password", "", "The etcd password to use for authentication.")

	flag.Parse()

	parsedEtcdEndpoints := strings.Split(*etcdEndpoints, ",")
	return &Config{
		InterfaceName: *ifaceName,
		PrivateIP:     *privateIP,
		PublicIP:      *publicIP,
		SubnetMask:    *subnetMask,
		ListenAddress: *laddr,
		ListenPort:    *lport,
		Prefix:        *prefix,
		LeaseTime:     *leaseTime,
		SyncInterval:  *syncInterval,
		Retries:       *retries,
		EnableCrypto:  *crypto,
		EtcdEndpoints: parsedEtcdEndpoints,
		EtcdUsername:  *etcdUsername,
		EtcdPassword:  *etcdPassword,
	}
}
