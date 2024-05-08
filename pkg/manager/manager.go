// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"github.com/onosproject/cco-mon/pkg/mho"
	nbi "github.com/onosproject/cco-mon/pkg/northbound"
	e2mho "github.com/onosproject/cco-mon/pkg/southbound/mho"
	e2rc "github.com/onosproject/cco-mon/pkg/southbound/rc"
	"github.com/onosproject/cco-mon/pkg/store/metrics"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-mho/pkg/store"

	"github.com/onosproject/cco-mon/pkg/southbound/e2policy"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	// "github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/cco-mon/pkg/config"
	"github.com/onosproject/cco-mon/pkg/controller"
	"github.com/onosproject/cco-mon/pkg/monitor"
	"github.com/onosproject/cco-mon/pkg/rnib"

	// mlbnbi "github.com/onosproject/cco-mon/pkg/northbound"
	ocnstorage "github.com/onosproject/cco-mon/pkg/store/store/ocn"
	paramstorage "github.com/onosproject/cco-mon/pkg/store/store/parameters"
	"github.com/onosproject/cco-mon/pkg/store/store/storage"
)

var log = logging.GetLogger("cco-mon", "manager")

// Config is a manager configuration
type Config struct {
	AppID        string
	CAPath       string
	KeyPath      string
	CertPath     string
	E2tAddress   string
	E2tPort      int
	TopoAddress  string
	TopoPort     int
	GRPCPort     int
	RicActionID  int32
	SMNameMHO    string
	SMVersionMHO string
	SMNameRC     string
	SMVersionRC  string
	// Ravi
	ConfigPath          string
	OverloadThreshold   int
	TargetLoadThreshold int
	E2tEndpoint         string
}

// NewManager generates the new CCO MON xAPP manager
func NewManager(parameters Config, flag bool) *Manager {

	//  TODO: fill the below missing dependecies
	// 1. numUEsMeasStore
	// 2. neighborMeasStore
	// 3. ocnStore
	// 4. parameters {} ---> { E2tEndpoint, }
	// 5. paramStore
	// 6. appCfg

	ueStore := store.NewStore()
	cellStore := store.NewStore()
	metricStore := metrics.NewStore()

	indCh := make(chan *mho.E2NodeIndication)
	ctrlReqChs := make(map[string]chan *e2api.ControlMessage)

	optionsKpi := e2mho.Options{
		AppID:       parameters.AppID,
		E2tAddress:  parameters.E2tAddress,
		E2tPort:     parameters.E2tPort,
		TopoAddress: parameters.TopoAddress,
		TopoPort:    parameters.TopoPort,
		SMName:      parameters.SMNameMHO,
		SMVersion:   parameters.SMVersionMHO,
	}

	e2ManagerKPI, err := e2mho.NewManager(optionsKpi, indCh, ctrlReqChs, ueStore, cellStore)
	if err != nil {
		log.Warn(err)
	}

	//e2 Manager for RC Subscriptions
	optionsRc := e2rc.Options{
		AppID:       parameters.AppID,
		E2tAddress:  parameters.E2tAddress,
		E2tPort:     parameters.E2tPort,
		TopoAddress: parameters.TopoAddress,
		TopoPort:    parameters.TopoPort,
		SMName:      parameters.SMNameRC,
		SMVersion:   parameters.SMVersionRC,
	}

	e2ManagerRC, err := e2rc.NewManager(optionsRc, metricStore)
	if err != nil {
		log.Warn(err)
	}

	numUEsMeasStore := storage.NewStore()
	neighborMeasStore := storage.NewStore()
	ocnStore := ocnstorage.NewStore()

	rnibOptions := rnib.Options{
		TopoAddress: parameters.TopoAddress,
		TopoPort:    parameters.TopoPort,
	}
	rnibHandler, err := rnib.NewHandler(rnibOptions)
	if err != nil {
		log.Error(err)
	}
	monitorHandler := monitor.NewHandler(rnibHandler, numUEsMeasStore, neighborMeasStore, ocnStore)

	//added by raavi

	appCfg, err := config.NewConfig(parameters.ConfigPath)
	if err != nil {
		log.Warn(err)
	}
	interval, err := appCfg.GetInterval(MLBAppIntervalPath)
	if err != nil {
		log.Warn("set interval to default interval - reason: %v", err)
		interval = MLBAppDefaultInterval
	}

	paramStore := paramstorage.NewStore()

	err = paramStore.Put(context.Background(), "interval", interval)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "delta_ocn", OCNDeltaFactor)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "overload_threshold", parameters.OverloadThreshold)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "target_threshold", parameters.TargetLoadThreshold)
	if err != nil {
		log.Error(err)
	}

	//e2ControlHandler := e2control.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion,
	//	AppID, parameters.E2tEndpoint)

	e2PolicyHandler := e2policy.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion, AppID, parameters.E2tEndpoint, rnibHandler)
	log.Info("Reached line no 155")
	//ctrlHandler := controller.NewHandler(e2ControlHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, ocnStore, paramStore)
	ctrlHandler := controller.NewHandler(e2PolicyHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, ocnStore, paramStore)
	//e2 Manager for MHO Subscriptions
	log.Info("Reached line no 159")

	// ocnStore := ocnstorage.NewStore()
	manager := &Manager{
		// config:       config,
		// parameters:   parameters,
		e2ManagerKpi: e2ManagerKPI,
		e2ManagerRc:  e2ManagerRC,
		mhoCtrl:      mho.NewController(indCh, ueStore, cellStore, flag),
		ueStore:      ueStore,
		cellStore:    cellStore,
		metricStore:  metricStore,
		ctrlReqChs:   ctrlReqChs,
		// ocnStore:     ocnStore,
		handlers: handlers{
			rnibHandler:    rnibHandler,
			monitorHandler: monitorHandler,
			//e2ControlHandler:  e2ControlHandler,
			e2PolicyHandler:   e2PolicyHandler,
			controllerHandler: ctrlHandler,
		},
		stores: stores{
			numUEsMeasStore:   numUEsMeasStore,
			neighborMeasStore: neighborMeasStore,
			ocnStore:          ocnStore,
			paramStore:        paramStore,
		},
		channels: channels{},
		configs: configs{
			appConfigParams: parameters,
			appConfig:       appCfg,
		},
		ocnStore: ocnStore,
	}
	return manager
}

// Manager is an abstract struct for manager
type Manager struct {
	config       Config
	e2ManagerKpi e2mho.Manager
	e2ManagerRc  e2rc.Manager
	mhoCtrl      *mho.Controller
	ueStore      store.Store
	cellStore    store.Store
	metricStore  metrics.Store
	ctrlReqChs   map[string]chan *e2api.ControlMessage
	ocnStore     ocnstorage.Store
	handlers     handlers
	stores       stores
	channels     channels
	configs      configs
}
type handlers struct {
	rnibHandler    rnib.Handler
	monitorHandler monitor.Handler
	//e2ControlHandler  e2control.Handler
	e2PolicyHandler   e2policy.Handler
	controllerHandler controller.Handler
}

type stores struct {
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

type channels struct {
}

type configs struct {
	appConfigParams Config
	appConfig       config.Config
}

func (m *Manager) Run() {
	err := m.start()
	if err != nil {
		log.Errorf("Error when starting CCO KPI Monitoring XApp: %v", err)
	}
}

func (m *Manager) startnbiserver() error {
	nbiManager := nbi.NewManager(m.configs.appConfigParams.GRPCPort, m.ueStore, m.metricStore, m.ocnStore)
	err := nbiManager.Start()
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (m *Manager) start() error {
	err := m.e2ManagerKpi.Start()
	if err != nil {
		log.Warn(err)
		return err
	}
	log.Info("Reached line no 259")

	// does this have to be syncronous? noo
	go func() {
		err = m.handlers.controllerHandler.Run(context.Background())
		if err != nil {
			log.Warn(err)
		}
	}() 

	//start mho manger to handle periodic measurement reports
	handleFlag := false
	go m.mhoCtrl.Run(context.Background(), &handleFlag)

	//start rc sm manager
	err = m.e2ManagerRc.Start()
	if err != nil {
		log.Warn(err)
		return err
	}

	//Start nbi server (grpc communication)
	err = m.startnbiserver()
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (m *Manager) Close() {
	log.Info("Closing Manager")
}

//Starts the mlb stuff ********************************************************************************

// // GetOcnStore returns Ocn store
// func (m *Manager1) GetOcnStore() ocnstorage.Store {

// 	log.Info("GetOcnStore")
// 	log.Infof("#manager.getocnstore.179new stores %#v",m.stores)
// 	log.Infof("#manager.getocnstore.180new ocn stores %#v",m.stores.ocnStore)
// 	return m.stores.ocnStore

// }

// // GetNumUEsStore returns NumUEsStore
// func (m *Manager1) GetNumUEsStore() storage.Store {
// 	log.Infof("GetNumUEsStore: %v", m.stores.numUEsMeasStore)
// 	return m.stores.numUEsMeasStore
// }

// // GetNeighborStore returns neighbor store
// func (m *Manager1) GetNeighborStore() storage.Store {
// 	log.Infof("GetNumUEsStore: %v", m.stores.neighborMeasStore)
// 	return m.stores.neighborMeasStore
// }

// // Start starts this app's manager
// func (m *Manager1) Start() error {
// 	/*err := m.startNorthboundServer()
// 	//if err != nil {
// 	//	log.Infof("error in starting nbi server %v", err)
// 		return err
// 	}*/
// 	log.Info("Hello World!")
// 	err := m.handlers.controllerHandler.Run(context.Background())
// 	return err
// }

// // Manager is a struct including this app's manager information and objects
// type Manager1 struct {
// 	handlers handlers
// 	stores   stores
// 	channels channels
// 	configs  configs
// }

// type handlers struct {
// 	rnibHandler    rnib.Handler
// 	monitorHandler monitor.Handler
// 	//e2ControlHandler  e2control.Handler
// 	e2PolicyHandler   e2policy.Handler
// 	controllerHandler controller.Handler
// }

// type stores struct {
// 	numUEsMeasStore   storage.Store
// 	neighborMeasStore storage.Store
// 	ocnStore          ocnstorage.Store //
// 	paramStore        paramstorage.Store
// }

// type channels struct {
// }

// type configs struct {
// 	appConfigParams AppParameters
// 	appConfig       config.Config
// }

// // AppParameters includes all application parameters coming from arguments when starting this app
// type AppParameters struct {
// 	CAPath              string
// 	KeyPath             string
// 	CertPath            string
// 	ConfigPath          string
// 	E2tEndpoint         string
// 	UENIBEndpoint       string
// 	GRPCPort            int
// 	RicActionID         int32
// 	OverloadThreshold   int
// 	TargetLoadThreshold int
// }

// // NewManager generates this application's manager
// func NewManager1(parameters AppParameters) *Manager1 {
// 	appCfg, err := config.NewConfig(parameters.ConfigPath)
// 	if err != nil {
// 		log.Warn(err)
// 	}
// 	interval, err := appCfg.GetInterval(MLBAppIntervalPath)
// 	if err != nil {
// 		log.Warn("set interval to default interval - reason: %v", err)
// 		interval = MLBAppDefaultInterval // tHe default is 10s
// 	}

// 	numUEsMeasStore := storage.NewStore() //
// 	neighborMeasStore := storage.NewStore()
// 	ocnStore := ocnstorage.NewStore()
// 	// New store comes as txPowerStore which also needs a txPowerRange struct (identiocal to qOffsetRange)
// 	paramStore := paramstorage.NewStore()
// 	// store for tx power
// 	err = paramStore.Put(context.Background(), "interval", interval) // For switch off tx Put is only required
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	err = paramStore.Put(context.Background(), "delta_ocn", OCNDeltaFactor)
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	err = paramStore.Put(context.Background(), "overload_threshold", parameters.OverloadThreshold)
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	err = paramStore.Put(context.Background(), "target_threshold", parameters.TargetLoadThreshold)
// 	if err != nil {
// 		log.Error(err)
// 	}

// 	rnibHandler, err := rnib.NewHandler()
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	monitorHandler := monitor.NewHandler(rnibHandler, numUEsMeasStore, neighborMeasStore, ocnStore)

// 	//e2ControlHandler := e2control.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion,
// 	//	AppID, parameters.E2tEndpoint)

// 	e2PolicyHandler := e2policy.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion, AppID, parameters.E2tEndpoint, rnibHandler)

// 	//ctrlHandler := controller.NewHandler(e2ControlHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, ocnStore, paramStore)
// 	ctrlHandler := controller.NewHandler(e2PolicyHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, ocnStore, paramStore)

// 	//ctrlHandler for txpowerdB how??

// 	return &Manager1{
// 		handlers: handlers{
// 			rnibHandler:    rnibHandler,
// 			monitorHandler: monitorHandler,
// 			//e2ControlHandler:  e2ControlHandler,
// 			e2PolicyHandler:   e2PolicyHandler,
// 			controllerHandler: ctrlHandler,
// 		},
// 		stores: stores{
// 			numUEsMeasStore:   numUEsMeasStore,
// 			neighborMeasStore: neighborMeasStore,
// 			ocnStore:          ocnStore,
// 			paramStore:        paramStore,
// 		},
// 		channels: channels{},
// 		configs: configs{
// 			appConfigParams: parameters,
// 			appConfig:       appCfg,
// 		},
// 	}
// }
