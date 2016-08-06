package config

import (
	"flag"
	"github.com/Supernomad/quantum/ecdh"
	"strconv"
	"strings"
	"time"
)

// Config handles marshalling user supplied configuration data
type Config struct {
	InterfaceName string

	PrivateIP string
	PublicIP  string

	PublicAddress string

	PrivateKey []byte
	PublicKey  []byte

	ListenAddress string
	ListenPort    int

	Prefix string

	LeaseTime       time.Duration
	SyncInterval    time.Duration
	RefreshInterval time.Duration

	TLSEnabled bool
	TLSCert    string
	TLSKey     string
	TLSCA      string

	Datastore string
	Endpoints []string
	Username  string
	Password  string
}

// New generates a new config object
func New() *Config {
	ifaceName := flag.String("interface-name", "quantum%d", "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")

	privateIP := flag.String("private-ip", "", "The private ip address of this node.")
	publicIP := flag.String("public-ip", "", "The public ip address of this node.")

	laddr := flag.String("listen-address", "0.0.0.0", "The ip address to listen on for forwarded packets.")
	lport := flag.Int("listen-port", 1099, "The ip port to listen on for forwarded packets.")

	prefix := flag.String("prefix", "quantum", "The etcd key that quantum information is stored under.")

	leaseTime := flag.Duration("lease-time", 300, "Lease time for the private ip address.")
	syncInterval := flag.Duration("sync-interval", 30, "The backend sync interval")
	refreshInterval := flag.Duration("refresh-interval", 60, "The backend lease refresh interval.")

	tlsCert := flag.String("tls-cert", "", "The client certificate to use for authentication with the backend datastore.")
	tlsKey := flag.String("tls-key", "", "The client key to use for authentication with the backend datastore.")
	tlsCA := flag.String("tls-ca-cert", "", "The CA certificate to authenticate the backend datastore.")

	datastore := flag.String("datastore", "etcd", "The datastore backend to use, either consul or etcd")
	endpoints := flag.String("endpoints", "127.0.0.1:2379", "The datastore endpoints to use, in a comma separated list.")
	username := flag.String("username", "", "The datastore username to use for authentication.")
	password := flag.String("password", "", "The datastore password to use for authentication.")

	flag.Parse()

	parsedEndpoints := strings.Split(*endpoints, ",")
	pubkey, privkey := ecdh.GenerateECKeyPair()

	tlsEnabled := false
	if *tlsCert != "" || *tlsKey != "" || *tlsCA != "" {
		tlsEnabled = true
	}

	return &Config{
		InterfaceName:   *ifaceName,
		PrivateIP:       *privateIP,
		PublicIP:        *publicIP,
		PublicAddress:   *publicIP + ":" + strconv.Itoa(*lport),
		PrivateKey:      privkey,
		PublicKey:       pubkey,
		ListenAddress:   *laddr,
		ListenPort:      *lport,
		Prefix:          *prefix,
		LeaseTime:       *leaseTime,
		SyncInterval:    *syncInterval,
		RefreshInterval: *refreshInterval,
		TLSCert:         *tlsCert,
		TLSKey:          *tlsKey,
		TLSCA:           *tlsCA,
		TLSEnabled:      tlsEnabled,
		Datastore:       *datastore,
		Endpoints:       parsedEndpoints,
		Username:        *username,
		Password:        *password,
	}
}
