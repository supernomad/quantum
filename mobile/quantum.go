package mobile

import (
	"sort"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/datastore"
	"github.com/supernomad/quantum/device"
	"github.com/supernomad/quantum/metric"
	"github.com/supernomad/quantum/plugin"
	"github.com/supernomad/quantum/rest"
	"github.com/supernomad/quantum/router"
	"github.com/supernomad/quantum/socket"
	"github.com/supernomad/quantum/worker"
)

func Init() string {
	log := common.NewLogger(common.NoopLogger)

	cfg, err := common.NewConfig(log)
	if err != nil {
		return err.Error()
	}

	store, err := datastore.New(datastore.ETCDDatastore, cfg)
	if err != nil {
		return err.Error()
	}

	err = store.Init()
	if err != nil {
		return err.Error()
	}

	incomingPlugins := make([]plugin.Plugin, len(cfg.Plugins))
	outgoingPlugins := make([]plugin.Plugin, len(cfg.Plugins))
	for i := 0; i < len(cfg.Plugins); i++ {
		plugin, err := plugin.New(cfg.Plugins[i], cfg)
		if err != nil {
			return err.Error()
		}
		outgoingPlugins[i] = plugin
		incomingPlugins[i] = plugin
	}

	sort.Sort(plugin.Sorter{Plugins: outgoingPlugins})
	sort.Sort(sort.Reverse(plugin.Sorter{Plugins: incomingPlugins}))

	dev, err := device.New(device.TUNDevice, cfg)
	if err != nil {
		return err.Error()
	}

	sock, err := socket.New(cfg.NetworkConfig.Backend, cfg)
	if err != nil {
		return err.Error()
	}

	aggregator := metric.New(cfg)

	api := rest.New(cfg, aggregator)

	rt := router.New(cfg, store)

	outgoing := worker.NewOutgoing(cfg, aggregator, rt, outgoingPlugins, dev, sock)
	incoming := worker.NewIncoming(cfg, aggregator, rt, incomingPlugins, dev, sock)

	api.Start()
	aggregator.Start()
	store.Start()

	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}
	return "Worked!"
}
