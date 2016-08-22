package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"github.com/go-playground/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config handles marshalling user supplied configuration data
type Config struct {
	InterfaceName string
	MachineID     string
	NumWorkers    int
	StatsWindow   time.Duration

	PrivateIP string
	PublicIP  string

	PublicAddress string

	PrivateKey []byte
	PublicKey  []byte

	ListenAddress string
	ListenPort    int

	Prefix   string
	ConfFile string
	DataDir  string
	LogDir   string

	SyncInterval    time.Duration
	RefreshInterval time.Duration

	TLSEnabled    bool
	TLSSkipVerify bool
	TLSCert       string
	TLSKey        string
	TLSCA         string

	Datastore   string
	endpoints   string
	Endpoints   []string
	AuthEnabled bool
	Username    string
	Password    string

	NetworkConfig *NetworkConfig
	notSet        map[string]bool
}

func (cfg *Config) handleDefaultString(name, def string) string {
	env := "QUANTUM_" + strings.ToUpper(strings.Replace(name, "-", "_", 10))
	output := os.Getenv(env)
	if output == "" {
		cfg.notSet[name] = true
		return def
	}
	return output
}

func (cfg *Config) handleDefaultInt(name string, def int) int {
	str := strconv.Itoa(def)
	output, err := strconv.Atoi(cfg.handleDefaultString(name, str))
	if err != nil {
		panic(err)
	}
	return output
}

func (cfg *Config) handleDefaultDuration(name string, def time.Duration) time.Duration {
	str := def.String()
	output, err := time.ParseDuration(cfg.handleDefaultString(name, str))
	if err != nil {
		panic(err)
	}
	return output
}

func (cfg *Config) handleDefaultBool(name string, def bool) bool {
	str := strconv.FormatBool(def)
	output, err := strconv.ParseBool(cfg.handleDefaultString(name, str))
	if err != nil {
		panic(err)
	}
	return output
}

func (cfg *Config) handleCli() {
	flag.StringVar(&cfg.ConfFile, "conf-file", cfg.handleDefaultString("conf-file", ""), "The json or yaml file to load configuration data from.")

	flag.StringVar(&cfg.InterfaceName, "interface-name", cfg.handleDefaultString("interface-name", "quantum%d"), "The name for the TUN interface that will be used for forwarding. Use %d to have the OS pick an available interface name.")
	flag.DurationVar(&cfg.StatsWindow, "stats-window", cfg.handleDefaultDuration("stats-window", 5), "The window of time to calculate bandwidth and packet per second information on.")

	flag.StringVar(&cfg.PrivateIP, "private-ip", cfg.handleDefaultString("private-ip", ""), "The private ip address of this node.")
	flag.StringVar(&cfg.PublicIP, "public-ip", cfg.handleDefaultString("public-ip", ""), "The public ip address of this node.")

	flag.StringVar(&cfg.ListenAddress, "listen-address", cfg.handleDefaultString("listen-address", "0.0.0.0"), "The ip address to listen on for forwarded packets.")
	flag.IntVar(&cfg.ListenPort, "listen-port", cfg.handleDefaultInt("listen-port", 1099), "The ip port to listen on for forwarded packets.")

	flag.StringVar(&cfg.Prefix, "prefix", cfg.handleDefaultString("prefix", "quantum"), "The etcd key that quantum information is stored under.")
	flag.StringVar(&cfg.DataDir, "data-dir", cfg.handleDefaultString("data-dir", "/var/lib/quantum"), "The data directory for quantum to use for persistent state.")
	flag.StringVar(&cfg.LogDir, "log-dir", cfg.handleDefaultString("log-dir", ""), "The log directory to write logs to, if this is ommited logs are written to stdout/stderr.")

	flag.DurationVar(&cfg.SyncInterval, "sync-interval", cfg.handleDefaultDuration("sync-interval", 30), "The backend sync interval")
	flag.DurationVar(&cfg.RefreshInterval, "refresh-interval", cfg.handleDefaultDuration("refresh-interval", 60), "The backend lease refresh interval.")

	flag.StringVar(&cfg.TLSCert, "tls-cert", cfg.handleDefaultString("tls-cert", ""), "The client certificate to use for authentication with the backend datastore.")
	flag.StringVar(&cfg.TLSKey, "tls-key", cfg.handleDefaultString("tls-key", ""), "The client key to use for authentication with the backend datastore.")
	flag.StringVar(&cfg.TLSCA, "tls-ca-cert", cfg.handleDefaultString("tls-ca-cert", ""), "The CA certificate to authenticate the backend datastore.")
	flag.BoolVar(&cfg.TLSSkipVerify, "tls-skip-verify", cfg.handleDefaultBool("tls-skip-verify", false), "The CA certificate to authenticate the backend datastore.")

	flag.StringVar(&cfg.Datastore, "datastore", cfg.handleDefaultString("datastore", "etcd"), "The datastore backend to use, either consul or etcd")
	flag.StringVar(&cfg.endpoints, "endpoints", cfg.handleDefaultString("endpoints", "127.0.0.1:2379"), "A comma delimited list of datastore endpoints to use.")
	flag.StringVar(&cfg.Username, "username", cfg.handleDefaultString("username", ""), "The datastore username to use for authentication.")
	flag.StringVar(&cfg.Password, "password", cfg.handleDefaultString("password", ""), "The datastore password to use for authentication.")

	flag.Parse()
}

func (cfg *Config) handleComputed() {
	cfg.Endpoints = strings.Split(cfg.endpoints, ",")

	pubkey, privkey := GenerateECKeyPair()
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

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores * 2)

	cfg.NumWorkers = cores
}

func (cfg *Config) parseFileData(data map[string]string) error {
	for k, v := range data {
		if _, ok := cfg.notSet[k]; ok {
			switch k {
			case "conf-file":
				cfg.ConfFile = v
			case "interface-name":
				cfg.InterfaceName = v
			case "private-ip":
				cfg.PrivateIP = v
			case "public-ip":
				cfg.PublicIP = v
			case "listen-address":
				cfg.ListenAddress = v
			case "listen-port":
				i, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				cfg.ListenPort = i
			case "prefix":
				cfg.Prefix = v
			case "data-dir":
				cfg.DataDir = v
			case "log-dir":
				cfg.LogDir = v
			case "stats-window":
				dur, err := time.ParseDuration(v)
				if err != nil {
					return err
				}
				cfg.StatsWindow = dur
			case "sync-interval":
				dur, err := time.ParseDuration(v)
				if err != nil {
					return err
				}
				cfg.SyncInterval = dur
			case "refresh-interval":
				dur, err := time.ParseDuration(v)
				if err != nil {
					return err
				}
				cfg.RefreshInterval = dur
			case "tls-skip-verify":
				b, err := strconv.ParseBool(v)
				if err != nil {
					return err
				}
				cfg.TLSSkipVerify = b
			case "tls-cert":
				cfg.TLSCert = v
			case "tls-key":
				cfg.TLSKey = v
			case "tls-ca-cert":
				cfg.TLSCA = v
			case "datastore":
				cfg.Datastore = v
			case "endpoints":
				cfg.endpoints = v
			case "username":
				cfg.Username = v
			case "password":
				cfg.Password = v
			}
		}
	}
	return nil
}

func (cfg *Config) handleFile() error {
	if cfg.ConfFile != "" {
		buf, err := ioutil.ReadFile(cfg.ConfFile)
		if err != nil {
			return err
		}

		data := make(map[string]string)
		ext := path.Ext(cfg.ConfFile)
		switch {
		case ".json" == ext:
			err = json.Unmarshal(buf, &data)
		case ".yaml" == ext || ".yml" == ext:
			err = yaml.Unmarshal(buf, &data)
		default:
			return errors.New("The configuration file is not in a supported format.")
		}

		return cfg.parseFileData(data)
	}
	return nil
}

// NewConfig generates a new config object
func NewConfig() (*Config, error) {
	cfg := &Config{notSet: make(map[string]bool)}
	cfg.handleCli()
	err := cfg.handleFile()
	if err != nil {
		log.Error("Error handling config:", err)
		return nil, err
	}
	cfg.handleComputed()
	return cfg, nil
}
