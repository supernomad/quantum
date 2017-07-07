// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Supernomad/quantum/crypto"
	"github.com/Supernomad/quantum/version"
	"github.com/vishvananda/netlink"
	"gopkg.in/yaml.v2"
)

const (
	envDatastorePrefix               = "QUANTUM_"
	linkLocal                        = "fe80::/10"
	defaultBackend                   = "udp"
	defaultNetwork                   = "10.99.0.0/16"
	defaultStaticRange               = "10.99.0.0/23"
	defaultLeaseTime   time.Duration = 48 * time.Hour
)

var (
	googleV4   = net.ParseIP("8.8.8.8")
	googleV6   = net.ParseIP("2001:4860:4860::8888")
	loopbackV4 = net.ParseIP("127.0.0.1")
	loopbackV6 = net.ParseIP("::1")
	allV4      = net.ParseIP("0.0.0.0")
	allV6      = net.ParseIP("::")
)

/*
Config struct that handles marshalling in user supplied configuration data from cli arguments, environment variables, and configuration file entries.

The user supplied configuration is processed via a structured hierarchy:
	- Cli arguments override both environment variables and configuration file entries.
	- Environment variables will override file entries but can be overridden by cli arguments.
	- Configuration file entries will be overridden by both environment variables and cli arguments.
	- Defaults are used in the case that the user does not define a configuration argument.

The only exceptions to the above are the two special cli argments '-h'|'--help' or '-v'|'--version' which will output usage information or version information respectively and then exit the application.
*/
type Config struct {
	ConfFile                 string            `internal:"false"  type:"string"    short:"c"    long:"conf-file"                   default:""                      description:"The configuration file to use to configure quantum."`
	DeviceName               string            `internal:"false"  type:"string"    short:"i"    long:"device-name"                 default:"quantum%d"             description:"The name to give the TUN device quantum uses, append '%d' to have auto incrementing names."`
	NumWorkers               int               `internal:"false"  type:"int"       short:"n"    long:"workers"                     default:"0"                     description:"The number of quantum workers to use, set to 0 for a worker per available cpu core."`
	PrivateIP                net.IP            `internal:"false"  type:"ip"        short:"ip"   long:"private-ip"                  default:""                      description:"The private ip address to assign this quantum instance."`
	ListenIP                 net.IP            `internal:"false"  type:"ip"        short:"lip"  long:"listen-ip"                   default:""                      description:"The local server ip to listen on, leave blank of automatic association."`
	ListenPort               int               `internal:"false"  type:"int"       short:"p"    long:"listen-port"                 default:"1099"                  description:"The local server port to listen on."`
	PublicIPv4               net.IP            `internal:"false"  type:"ip"        short:"4"    long:"public-v4"                   default:""                      description:"The public ipv4 address to associate with this quantum instance, leave blank for automatic association."`
	DisableIPv4              bool              `internal:"false"  type:"bool"      short:"d4"   long:"disable-v4"                  default:"false"                 description:"Whether or not to disable public ipv4 auto addressing. Use this if you know the server doesn't have public ipv4 addressing."`
	PublicIPv6               net.IP            `internal:"false"  type:"ip"        short:"6"    long:"public-v6"                   default:""                      description:"The public ipv6 address to associate with this quantum instance, leave blank for automatic association."`
	DisableIPv6              bool              `internal:"false"  type:"bool"      short:"d6"   long:"disable-v6"                  default:"false"                 description:"Whether or not to disable public ipv6 auto addressing. Use this if you know the server doesn't have public ipv6 addressing."`
	DataDir                  string            `internal:"false"  type:"string"    short:"d"    long:"data-dir"                    default:"/var/lib/quantum"      description:"The directory to store local quantum state to."`
	PidFile                  string            `internal:"false"  type:"string"    short:"pf"   long:"pid-file"                    default:"/var/run/quantum.pid"  description:"The pid file to use for tracking rolling restarts."`
	Plugins                  []string          `internal:"false"  type:"list"      short:"x"    long:"plugins"                     default:""                      description:"The plugins supported by this node."`
	DatastorePrefix          string            `internal:"false"  type:"string"    short:"pr"   long:"datastore-prefix"            default:"quantum"               description:"The prefix to store quantum configuration data under in the key/value datastore."`
	DatastoreSyncInterval    time.Duration     `internal:"false"  type:"duration"  short:"si"   long:"datastore-sync-interval"     default:"60s"                   description:"The interval of full datastore syncs."`
	DatastoreRefreshInterval time.Duration     `internal:"false"  type:"duration"  short:"ri"   long:"datastore-refresh-interval"  default:"120s"                  description:"The interval of dhcp lease refreshes with the datastore."`
	DatastoreEndpoints       []string          `internal:"false"  type:"list"      short:"e"    long:"datastore-endpoints"         default:"127.0.0.1:2379"        description:"A comma delimited list of key/value datastore endpoints, in 'IPADDR:PORT' syntax."`
	DatastoreUsername        string            `internal:"false"  type:"string"    short:"u"    long:"datastore-username"          default:""                      description:"The username to use for authentication with the datastore."`
	DatastorePassword        string            `internal:"false"  type:"string"    short:"pw"   long:"datastore-password"          default:""                      description:"The password to use for authentication with the datastore."`
	DatastoreTLSSkipVerify   bool              `internal:"false"  type:"bool"      short:"tsv"  long:"datastore-tls-skip-verify"   default:"false"                 description:"Whether or not to authenticate the TLS certificates of the key/value datastore."`
	DatastoreTLSCA           string            `internal:"false"  type:"string"    short:"tca"  long:"datastore-tls-ca-cert"       default:""                      description:"The TLS CA certificate to authenticate the TLS certificates of the key/value datastore certificates."`
	DatastoreTLSCert         string            `internal:"false"  type:"string"    short:"tc"   long:"datastore-tls-cert"          default:""                      description:"The TLS client certificate to use to authenticate with the key/value datastore."`
	DatastoreTLSKey          string            `internal:"false"  type:"string"    short:"tk"   long:"datastore-tls-key"           default:""                      description:"The TLS client key to use to authenticate with the key/value datastore."`
	DTLSSkipVerify           bool              `internal:"false"  type:"bool"      short:"dtsv" long:"dtls-skip-verify"            default:"false"                 description:"Whether or not to authenticate the DTLS certificates when using the DTLS backend."`
	DTLSCA                   string            `internal:"false"  type:"string"    short:"dtca" long:"dtls-ca-cert"                default:""                      description:"The DTLS CA certificate to authenticate the DTLS certificates when using the DTLS backend."`
	DTLSCert                 string            `internal:"false"  type:"string"    short:"dtc"  long:"dtls-cert"                   default:""                      description:"The DTLS client certificate to use to authenticate when using the DTLS backend."`
	DTLSKey                  string            `internal:"false"  type:"string"    short:"dtk"  long:"dtls-key"                    default:""                      description:"The DTLS client key to use to authenticate when using the DTLS backend."`
	StatsRoute               string            `internal:"false"  type:"string"    short:"sr"   long:"stats-route"                 default:"/stats"                description:"The api route to serve statistics data from."`
	StatsAddress             string            `internal:"false"  type:"string"    short:"sa"   long:"stats-address"               default:"0.0.0.0"               description:"The api server address."`
	StatsPort                int               `internal:"false"  type:"int"       short:"sp"   long:"stats-port"                  default:"1099"                  description:"The api server port."`
	Network                  string            `internal:"false"  type:"string"    short:"nw"   long:"network"                     default:"10.99.0.0/16"          description:"The network, in CIDR notation, to use for the entire quantum cluster."`
	NetworkStaticRange       string            `internal:"false"  type:"string"    short:"nr"   long:"network-static-range"        default:"10.99.0.0/23"          description:"The reserved subnet, in CIDR notatio, within the network to use for static ip address assignments."`
	NetworkBackend           string            `internal:"false"  type:"string"    short:"nb"   long:"network-backend"             default:"udp"                   description:"The network backend to set in the datastore, if nothing already exists in the network configuration."`
	NetworkLeaseTime         time.Duration     `internal:"false"  type:"duration"  short:"nl"   long:"network-lease-time"          default:"48h"                   description:"The lease time for DHCP assigned addresses within the quantum cluster."`
	PublicKey                []byte            `internal:"true"` // The public key to use with the encryption plugin.
	PrivateKey               []byte            `internal:"true"` // The private key to use with the encryption plugin.
	PublicSalt               []byte            `internal:"true"` // The public salt to use with the encryption plugin.
	PrivateSalt              []byte            `internal:"true"` // The private salt to use with the encryption plugin.
	Salt                     []byte            `internal:"true"` // The salt to use with the encryption plugin.
	RealDeviceName           string            `internal:"true"` // Used when a rolling restart is triggered to find the correct tun interface
	ReuseFDS                 bool              `internal:"true"` // Used when a rolling restart is triggered which forces quantum to reuse the passed in socket/tun fds
	MachineID                string            `internal:"true"` // The generated machine id for this node
	AuthEnabled              bool              `internal:"true"` // Whether or not datastore authentication is enabled (toggled by setting username/password)
	TLSEnabled               bool              `internal:"true"` // Whether or not tls with the datastore is enabled (toggled by setting the tls parameters at run time)
	IsIPv4Enabled            bool              `internal:"true"` // Whether or not quantum has determined that this node is ipv4 capable
	IsIPv6Enabled            bool              `internal:"true"` // Whether or not quantum has determined that this node is ipv6 capable
	ListenAddr               syscall.Sockaddr  `internal:"true"` // The commputed Sockaddr object to bind the underlying udp sockets to
	NetworkConfig            *NetworkConfig    `internal:"true"` // The network config detemined by existence of the object in etcd
	Log                      *Logger           `internal:"true"` // The internal Logger to use
	fileData                 map[string]string `internal:"true"` // An internal map of data representing a passed in configuration file
}

func (cfg *Config) cliArg(short, long string, isFlag bool) (string, bool) {
	for i, arg := range os.Args {
		if arg == "-"+short ||
			arg == "--"+short ||
			arg == "-"+long ||
			arg == "--"+long {
			if !isFlag {
				return os.Args[i+1], true
			}
			return "true", true
		}
	}
	return "", false
}

func (cfg *Config) envArg(long string) (string, bool) {
	env := envDatastorePrefix + strings.ToUpper(strings.Replace(long, "-", "_", 10))
	output := os.Getenv(env)
	if output == "" {
		return output, false
	}
	return output, true
}

func (cfg *Config) fileArg(long string) (string, bool) {
	if cfg.fileData == nil {
		return "", false
	}
	value, ok := cfg.fileData[long]
	return value, ok
}

func (cfg *Config) usage(exit bool) {
	cfg.Log.Plain.Println("Usage of quantum:")
	st := reflect.TypeOf(*cfg)

	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		internal, fieldType, short, long, def, description := cfg.parseField(field.Tag)
		if internal == "true" {
			continue
		}

		cfg.Log.Plain.Printf("\t-%s|--%s  (%s)\n", short, long, fieldType)
		cfg.Log.Plain.Printf("\t\t%s (default: '%s')\n", description, def)
	}

	if exit {
		os.Exit(1)
	}
}

func (cfg *Config) version(exit bool) {
	cfg.Log.Plain.Printf("quantum: v%s\n", version.Version())

	if exit {
		os.Exit(0)
	}
}

func (cfg *Config) parseFile() error {
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
			err = errors.New("the supplied configuration file is not in a supported format, quantum only supports 'json', or 'yaml' configuration files")
		}

		if err != nil {
			return err
		}

		cfg.fileData = data
	}
	return nil
}

func (cfg *Config) parseField(tag reflect.StructTag) (internal, fieldType, short, long, def, description string) {
	internal = tag.Get("internal")
	fieldType = tag.Get("type")
	short = tag.Get("short")
	long = tag.Get("long")
	def = tag.Get("default")
	description = tag.Get("description")
	return
}

func (cfg *Config) parseSpecial(exit bool) {
	for _, arg := range os.Args {
		switch {
		case arg == "-h" || arg == "--h" || arg == "-help" || arg == "--help":
			cfg.usage(exit)
		case arg == "-v" || arg == "--v" || arg == "-version" || arg == "--version":
			cfg.version(exit)
		}
	}
}

func (cfg *Config) parseArgs() error {
	st := reflect.TypeOf(*cfg)
	sv := reflect.ValueOf(cfg).Elem()

	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		fieldValue := sv.Field(i)
		internal, fieldType, short, long, def, _ := cfg.parseField(field.Tag)

		if internal == "true" || !fieldValue.CanSet() {
			continue
		}

		var raw string
		if value, ok := cfg.cliArg(short, long, fieldType == "bool"); ok {
			raw = value
		} else if value, ok := cfg.envArg(long); ok {
			raw = value
		} else if value, ok := cfg.fileArg(long); ok {
			raw = value
		} else {
			raw = def
		}

		switch fieldType {
		case "int":
			i, err := strconv.Atoi(raw)
			if err != nil {
				return errors.New("error parsing value for '" + long + "' got, '" + raw + "', expected an 'int'")
			}
			fieldValue.Set(reflect.ValueOf(i))
		case "duration":
			dur, err := time.ParseDuration(raw)
			if err != nil {
				return errors.New("error parsing value for '" + long + "' got, '" + raw + "', expected a 'duration' for example: '10s' or '2d'")
			}
			fieldValue.Set(reflect.ValueOf(dur))
		case "ip":
			ip := net.ParseIP(raw)
			if ip == nil && raw != "" {
				return errors.New("error parsing value for '" + long + "' got, '" + raw + "', expected an 'ip' for example: '10.0.0.1' or 'fd42:dead:beef::1'")
			}
			fieldValue.Set(reflect.ValueOf(ip))
		case "bool":
			b, err := strconv.ParseBool(raw)
			if err != nil {
				return errors.New("error parsing value for '" + long + "' got, '" + raw + "', expected a 'bool'")
			}
			fieldValue.Set(reflect.ValueOf(b))
		case "list":
			if raw != "" {
				list := strings.Split(raw, ",")
				fieldValue.Set(reflect.ValueOf(list))
			} else {
				fieldValue.Set(reflect.ValueOf([]string{}))
			}
		case "string":
			fieldValue.Set(reflect.ValueOf(raw))
		default:
			return errors.New("build error unknown configuration type")
		}

		if field.Name == "ConfFile" {
			if err := cfg.parseFile(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cfg *Config) computeArgs() error {
	if (cfg.DatastoreTLSCert != "" && cfg.DatastoreTLSKey != "") || cfg.DatastoreTLSCA != "" {
		cfg.TLSEnabled = true
	}

	if cfg.DatastoreUsername != "" {
		cfg.AuthEnabled = true
	}

	if numCPU := runtime.NumCPU(); cfg.NumWorkers == 0 || cfg.NumWorkers > numCPU {
		cfg.NumWorkers = numCPU
	}

	os.MkdirAll(cfg.DataDir, os.ModeDir)
	os.MkdirAll(path.Dir(cfg.PidFile), os.ModeDir)

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

	cfg.RealDeviceName = os.Getenv(RealDeviceNameEnv)
	if cfg.RealDeviceName != "" {
		cfg.ReuseFDS = true
	}

	if StringInSlice("encryption", cfg.Plugins) {
		pub, priv := crypto.GenerateECKeyPair()
		pubSalt, privSalt := crypto.GenerateECKeyPair()

		cfg.PublicKey = pub
		cfg.PrivateKey = priv
		cfg.PublicSalt = pubSalt
		cfg.PrivateSalt = privSalt
	}

	DefaultNetworkConfig := &NetworkConfig{
		Backend:     cfg.NetworkBackend,
		Network:     cfg.Network,
		StaticRange: cfg.NetworkStaticRange,
		LeaseTime:   cfg.NetworkLeaseTime,
	}

	if DefaultNetworkConfig.Backend == "" {
		cfg.Log.Warn.Println("[CONFIG]", "Using default network backend:", defaultBackend)
		DefaultNetworkConfig.Backend = defaultBackend
	}

	if DefaultNetworkConfig.Network == "" {
		cfg.Log.Warn.Println("[CONFIG]", "Using default network:", defaultNetwork)
		cfg.Log.Warn.Println("[CONFIG]", "Using default network static range:", defaultStaticRange)
		DefaultNetworkConfig.Network = defaultNetwork
		DefaultNetworkConfig.StaticRange = defaultStaticRange
	}

	if DefaultNetworkConfig.LeaseTime == 0 {
		cfg.Log.Warn.Println("[CONFIG]", "Using default network lease time:", defaultLeaseTime.String())
		DefaultNetworkConfig.LeaseTime = defaultLeaseTime
	}

	baseIP, ipnet, err := net.ParseCIDR(DefaultNetworkConfig.Network)
	if err != nil {
		return err
	}

	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	if DefaultNetworkConfig.StaticRange != "" {
		staticBase, staticNet, err := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
		if err != nil {
			return err
		} else if !ipnet.Contains(staticBase) {
			return errors.New("network configuration has staticRange defined but the range does not exist in the configured network")
		}

		DefaultNetworkConfig.StaticNet = staticNet
	}

	cfg.NetworkConfig = DefaultNetworkConfig

	if cfg.PublicIPv4 == nil && !cfg.DisableIPv4 {
		routes, err := netlink.RouteGet(googleV4)
		if err != nil {
			return errors.New("error retrieving ipv4 route information, check to ensure valid network configuration exists on at the very least the loopback interface")
		}
		if !ArrayEquals(routes[0].Src, loopbackV4) {
			cfg.PublicIPv4 = routes[0].Src
			cfg.IsIPv4Enabled = true
		}
	} else if cfg.DisableIPv4 {
		cfg.IsIPv4Enabled = false
	} else {
		cfg.IsIPv4Enabled = true
	}

	if cfg.PublicIPv6 == nil && !cfg.DisableIPv6 {
		routes, err := netlink.RouteGet(googleV6)
		if err != nil {
			return errors.New("error retrieving ipv6 route information, check to ensure valid network configuration exists on at the very least the loopback interface")
		}
		_, ipNet, err := net.ParseCIDR(linkLocal)
		if err != nil {
			return errors.New("error parsing ipv6 linkLocal addressing")
		}
		if !ArrayEquals(routes[0].Src, loopbackV6) && !ipNet.Contains(routes[0].Src) {
			cfg.PublicIPv6 = routes[0].Src
			cfg.IsIPv6Enabled = true
		}
	} else if cfg.DisableIPv6 {
		cfg.IsIPv6Enabled = false
	} else {
		cfg.IsIPv6Enabled = true
	}

	if cfg.ListenIP == nil {
		switch {
		case cfg.IsIPv4Enabled && cfg.IsIPv6Enabled:
			fallthrough
		case cfg.IsIPv6Enabled:
			sa := &syscall.SockaddrInet6{Port: cfg.ListenPort}
			copy(sa.Addr[:], allV6.To16()[:])
			cfg.ListenAddr = sa
		case cfg.IsIPv4Enabled:
			sa := &syscall.SockaddrInet4{Port: cfg.ListenPort}
			copy(sa.Addr[:], allV4.To4()[:])
			cfg.ListenAddr = sa
		default:
			return errors.New("an impossible situation occurred, neither ipv4 or ipv6 is available, check your networking configuration you must have public internet access to use automatic configuration")
		}
	} else if addr := cfg.ListenIP.To4(); addr != nil {
		sa := &syscall.SockaddrInet4{Port: cfg.ListenPort}
		copy(sa.Addr[:], addr[:])
		cfg.ListenAddr = sa
	} else if addr := cfg.ListenIP.To16(); addr != nil {
		sa := &syscall.SockaddrInet6{Port: cfg.ListenPort}
		copy(sa.Addr[:], addr[:])
		cfg.ListenAddr = sa
	} else {
		return errors.New("an impossible situation occurred, neither ipv4 or ipv6 is available, check your networking configuration you must have public internet access to use automatic configuration")
	}

	pid := os.Getpid()
	return ioutil.WriteFile(cfg.PidFile, []byte(strconv.Itoa(pid)), os.ModePerm)
}

// NewConfig creates a new Config struct based on user supplied input.
func NewConfig(log *Logger) (*Config, error) {
	cfg := &Config{
		Log: log,
	}

	// Handle the help and version commands if the exist
	cfg.parseSpecial(true)

	// Handle parsing user supplied configuration data
	if err := cfg.parseArgs(); err != nil {
		return nil, err
	}

	// Compute internal configuration based on the user supplied configuration data
	if err := cfg.computeArgs(); err != nil {
		return nil, err
	}

	return cfg, nil
}
