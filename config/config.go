package config

import (
	"flag"
	"strings"
	"time"
)

// Config handles marshalling user supplied configuration data
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

	TlsCert string
	TlsKey  string
	TlsCA   string

	Datastore string
	Endpoints []string
	Username  string
	Password  string
}

// New generates a new config object
func New() *Config {
	ifaceName := flag.String("interface-name", "quantum", "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")
	privateIP := flag.String("private-ip", "", "The private ip address of this node.")
	publicIP := flag.String("public-ip", "", "The public ip address of this node.")
	subnetMask := flag.String("subnet-mask", "16", "The subnet mask in bit width format")

	laddr := flag.String("listen-address", "0.0.0.0", "The ip address to listen on for forwarded packets.")
	lport := flag.Int("listen-port", 1099, "The ip port to listen on for forwarded packets.")

	prefix := flag.String("prefix", "quantum", "The etcd key that quantum information is stored under.")
	leaseTime := flag.Duration("lease-time", 300, "Lease time for the private ip address.")
	syncInterval := flag.Duration("sync-interval", 30, "The backend sync interval")
	retries := flag.Duration("retries", 5, "The number of times to retry aquiring the private ip address lease.")
	crypto := flag.Bool("crypto", true, "Whether or not to encrypt data sent and recieved, by this node, to and from the rest of the cluster.")

	tlsCert := flag.String("tls-cert", "", "The client certificate to use for authentication with the backend datastore.")
	tlsKey := flag.String("tls-key", "", "The client key to use for authentication with the backend datastore.")
	tlsCA := flag.String("tls-ca-cert", "", "The CA certificate to authenticate the backend datastore.")

	datastore := flag.String("datastore", "etcd", "The datastore backend to use, either consul or etcd")
	endpoints := flag.String("endpoints", "127.0.0.1:2379", "The datastore endpoints to use, in a comma separated list.")
	username := flag.String("username", "", "The datastore username to use for authentication.")
	password := flag.String("password", "", "The datastore password to use for authentication.")

	flag.Parse()

	parsedEndpoints := strings.Split(*endpoints, ",")
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
		TlsCert:       *tlsCert,
		TlsKey:        *tlsKey,
		TlsCA:         *tlsCA,
		Datastore:     *datastore,
		Endpoints:     parsedEndpoints,
		Username:      *username,
		Password:      *password,
	}
}
