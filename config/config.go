package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"github.com/Supernomad/quantum/ecdh"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Config handles marshalling user supplied configuration data
type Config struct {
	InterfaceName string
	MachineID     string

	PrivateIP string
	PublicIP  string

	PublicAddress string

	PrivateKey []byte
	PublicKey  []byte

	ListenAddress string
	ListenPort    int

	Prefix  string
	DataDir string
	LogDir  string

	SyncInterval    time.Duration
	RefreshInterval time.Duration

	TLSEnabled bool
	TLSCert    string
	TLSKey     string
	TLSCA      string

	Datastore   string
	endpoints   string
	Endpoints   []string
	AuthEnabled bool
	Username    string
	Password    string
}

func handleDefaultString(name, def string) string {
	env := "QUANTUM_" + strings.ToUpper(strings.Replace(name, "-", "_", 10))
	output := os.Getenv(env)
	if output == "" {
		return def
	}
	return output
}

func handleDefaultInt(name string, def int) int {
	str := strconv.Itoa(def)
	output, err := strconv.Atoi(handleDefaultString(name, str))
	if err != nil {
		panic(err)
	}
	return output
}

func handleDefaultDuration(name string, def time.Duration) time.Duration {
	str := def.String()
	output, err := time.ParseDuration(handleDefaultString(name, str))
	if err != nil {
		panic(err)
	}
	return output
}

func (cfg *Config) handleCli() {
	flag.StringVar(&cfg.InterfaceName, "interface-name", handleDefaultString("interface-name", "quantum%d"), "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")

	flag.StringVar(&cfg.PrivateIP, "private-ip", handleDefaultString("private-ip", ""), "The private ip address of this node.")
	flag.StringVar(&cfg.PublicIP, "public-ip", handleDefaultString("public-ip", ""), "The public ip address of this node.")

	flag.StringVar(&cfg.ListenAddress, "listen-address", handleDefaultString("listen-address", "0.0.0.0"), "The ip address to listen on for forwarded packets.")
	flag.IntVar(&cfg.ListenPort, "listen-port", handleDefaultInt("listen-port", 1099), "The ip port to listen on for forwarded packets.")

	flag.StringVar(&cfg.Prefix, "prefix", handleDefaultString("prefix", "quantum"), "The etcd key that quantum information is stored under.")
	flag.StringVar(&cfg.DataDir, "data-dir", handleDefaultString("data-dir", "/var/lib/quantum"), "The data directory for quantum to use for persistent state.")
	flag.StringVar(&cfg.LogDir, "log-dir", handleDefaultString("log-dir", ""), "The log directory to write logs to, if this is ommited logs are written to stdout/stderr.")

	flag.DurationVar(&cfg.SyncInterval, "sync-interval", handleDefaultDuration("sync-interval", 30), "The backend sync interval")
	flag.DurationVar(&cfg.RefreshInterval, "refresh-interval", handleDefaultDuration("refresh-interval", 60), "The backend lease refresh interval.")

	flag.StringVar(&cfg.TLSCert, "tls-cert", handleDefaultString("tls-cert", ""), "The client certificate to use for authentication with the backend datastore.")
	flag.StringVar(&cfg.TLSKey, "tls-key", handleDefaultString("tls-key", ""), "The client key to use for authentication with the backend datastore.")
	flag.StringVar(&cfg.TLSCA, "tls-ca-cert", handleDefaultString("tls-ca-cert", ""), "The CA certificate to authenticate the backend datastore.")

	flag.StringVar(&cfg.Datastore, "datastore", handleDefaultString("datastore", "etcd"), "The datastore backend to use, either consul or etcd")

	flag.StringVar(&cfg.endpoints, "endpoints", handleDefaultString("endpoints", "127.0.0.1:2379"), "A comma delimited list of datastore endpoints to use.")
	flag.StringVar(&cfg.Username, "username", handleDefaultString("username", ""), "The datastore username to use for authentication.")
	flag.StringVar(&cfg.Password, "password", handleDefaultString("password", ""), "The datastore password to use for authentication.")

	flag.Parse()
}

func (cfg *Config) handleComputed() {
	cfg.Endpoints = strings.Split(cfg.endpoints, ",")

	pubkey, privkey := ecdh.GenerateECKeyPair()
	cfg.PublicKey = pubkey
	cfg.PrivateKey = privkey

	if (cfg.TLSCert != "" && cfg.TLSKey != "") || cfg.TLSCA != "" {
		cfg.TLSEnabled = true
	}

	if cfg.Username != "" {
		cfg.AuthEnabled = true
	}

	os.MkdirAll(cfg.DataDir, os.ModeDir)
	machineID := make([]byte, 32)
	machineIDPath := path.Join(cfg.DataDir, "machine-id")
	if _, err := os.Stat(machineIDPath); os.IsNotExist(err) {
		rand.Read(machineID)
		ioutil.WriteFile(machineIDPath, machineID, os.ModePerm)
	} else {
		buf, _ := ioutil.ReadFile(machineIDPath)
		machineID = buf
	}
	cfg.MachineID = hex.EncodeToString(machineID)

	cfg.PublicAddress = cfg.PublicIP + ":" + strconv.Itoa(cfg.ListenPort)
}

// New generates a new config object
func New() *Config {
	cfg := &Config{}
	cfg.handleCli()
	cfg.handleComputed()
	return cfg
}
