// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"

	"github.com/onosproject/cco-mon/pkg/manager"
	"github.com/onosproject/cco-mon/pkg/utils"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/logging"
)

var log = logging.GetLogger("cco-mon")

func main() {

	log.SetLevel(logging.DebugLevel)

	err := utils.SetupTestSuite()
	if err != nil {
		log.Fatal(err)
	}

	caPath := flag.String("caPath", "/tmp/tls.cacrt", "path to CA certificate")
	keyPath := flag.String("keyPath", "/tmp/tls.key", "path to client private key")
	certPath := flag.String("certPath", "/tmp/tls.crt", "path to client certificate")
	e2tAddress := flag.String("e2tAddress", "onos-e2t", "E2T service address")
	e2tPort := flag.Int("e2tPort", 5150, "E2T service port")
	topoAddress := flag.String("topoAddress", "onos-topo", "Topology service address")
	topoPort := flag.Int("topoPort", 5150, "Topology service port")
	ricActionID := flag.Int("ricActionID", 10, "RIC Action ID in E2 message")
	grpcPort := flag.Int("grpcPort", 5150, "grpc Port number")
	smNamemho := flag.String("smNamemho", "oran-e2sm-mho", "Service model name in RAN function description")
	smVersionmho := flag.String("smVersionmho", "v2", "Service model version in RAN function description")
	smNamerc := flag.String("smNamerc", "oran-e2sm-rc", "Service model name in RAN function description")
	smVersionrc := flag.String("smVersionrc", "v1", "Service model version in RAN function description")
	AppID := flag.String("AppID", "cco-mon", "Application ID")
//For MLB
	configPath := flag.String("configPath", "/etc/onos/config/config.json", "path to config.json file")
	e2tEndpoint := flag.String("e2tEndpoint", "onos-e2t:5150", "E2T service endpoint")
	// uenibEndpoint := flag.String("uenibEndpoint", "onos-uenib:5150", "UENIB service endpoint")
	overloadThreshold := flag.Int("overloadThreshold", 20, "Overload threshold")
	targetLoadThreshold := flag.Int("targetLoadThreshold", 0, "Target load threshold")



	ready := make(chan bool)

	flag.Parse()

	_, err = certs.HandleCertPaths(*caPath, *keyPath, *certPath, true)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Starting CCO KPI Monitoring XApp")
	cfg := manager.Config{
		AppID:        *AppID,
		CAPath:       *caPath,
		KeyPath:      *keyPath,
		CertPath:     *certPath,
		E2tAddress:   *e2tAddress,
		E2tPort:      *e2tPort,
		TopoAddress:  *topoAddress,
		TopoPort:     *topoPort,
		GRPCPort:     *grpcPort,
		RicActionID:  int32(*ricActionID),
		SMNameMHO:    *smNamemho,
		SMVersionMHO: *smVersionmho,
		SMNameRC:     *smNamerc,
		SMVersionRC:  *smVersionrc,
		E2tEndpoint:  *e2tEndpoint,
		// UENIBEndpoint:       *uenibEndpoint,
		ConfigPath:          *configPath,
		OverloadThreshold:   *overloadThreshold,
		TargetLoadThreshold: *targetLoadThreshold,


	}

	mgr := manager.NewManager(cfg, false)
	// mgr.Run()

	// For mLB
	log.Info("Starting onos-mlb")
	log.Infof("RIC action ID is %v",ricActionID)


	// appConfParams := manager.AppParameters{
	// 	CAPath:              *caPath,
	// 	KeyPath:             *keyPath,
	// 	CertPath:            *certPath,
	// 	ConfigPath:          *configPath,
	// 	E2tEndpoint:         *e2tEndpoint,
	// 	UENIBEndpoint:       *uenibEndpoint,
	// 	GRPCPort:            *grpcPort,
	// 	RicActionID:         int32(*ricActionID),
	// 	OverloadThreshold:   *overloadThreshold,
	// 	TargetLoadThreshold: *targetLoadThreshold,
	// }

	// appMgr := manager.NewManager1(appConfParams) // initializes my Manager with the parameters

	// err = appMgr.Start()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// appMgr.Run()

	// Run both managers concurrently in their own goroutines
	// wait this is old? yes yes but without new grpc thing 
    log.Info("OG Manager")
    mgr.Run()

	// this is what we worked on yea? this is 
    // go func() {
	// 	log.Info("*********Starting onos-mlb")

    //     if err := appMgr.Start(); err != nil {
    //         log.Error("Error starting onos-mlb manager:", err)
    //     }
    // }()


	<-ready
}


// package main

// import (
// 	"flag"
// 	"github.com/onosproject/cco-mon/pkg/manager"
// 	"github.com/onosproject/cco-mon/pkg/utils"
// 	"github.com/onosproject/onos-lib-go/pkg/certs"
// 	"github.com/onosproject/onos-lib-go/pkg/logging"
// )

// var log = logging.GetLogger("cco-mon")

// func main() {
// 	log.SetLevel(logging.DebugLevel)

// 	err := utils.SetupTestSuite()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Flag definitions
// 	caPath := flag.String("caPath", "/tmp/tls.cacrt", "path to CA certificate")
// 	keyPath := flag.String("keyPath", "/tmp/tls.key", "path to client private key")
// 	certPath := flag.String("certPath", "/tmp/tls.crt", "path to client certificate")
// 	e2tAddress := flag.String("e2tAddress", "onos-e2t", "E2T service address")
// 	e2tPort := flag.Int("e2tPort", 5150, "E2T service port")
// 	topoAddress := flag.String("topoAddress", "onos-topo", "Topology service address")
// 	topoPort := flag.Int("topoPort", 5150, "Topology service port")
// 	ricActionID := flag.Int("ricActionID", 10, "RIC Action ID in E2 message")
// 	grpcPort := flag.Int("grpcPort", 5150, "GRPC Port number")
// 	smNamemho := flag.String("smNamemho", "oran-e2sm-mho", "Service model name for MHO")
// 	smVersionmho := flag.String("smVersionmho", "v2", "Service model version for MHO")
// 	smNamerc := flag.String("smNamerc", "oran-e2sm-rc", "Service model name for RC")
// 	smVersionrc := flag.String("smVersionrc", "v1", "Service model version for RC")
// 	AppID := flag.String("AppID", "cco-mon", "Application ID")
// 	// For MLB
// 	// configPath := flag.String("configPath", "/etc/onos/config/config.json", "Path to config.json file")
// 	// e2tEndpoint := flag.String("e2tEndpoint", "onos-e2t:5150", "E2T service endpoint")
// 	// uenibEndpoint := flag.String("uenibEndpoint", "onos-uenib:5150", "UENIB service endpoint")
// 	overloadThreshold := flag.Int("overloadThreshold", 20, "Overload threshold")
// 	targetLoadThreshold := flag.Int("targetLoadThreshold", 0, "Target load threshold")

// 	flag.Parse()

// 	// Handling TLS certificates
// 	_, err = certs.HandleCertPaths(*caPath, *keyPath, *certPath, true)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Consolidated Configuration
// 	cfg := manager.Config{
// 		AppID:               *AppID,
// 		CAPath:              *caPath,
// 		KeyPath:             *keyPath,
// 		CertPath:            *certPath,
// 		E2tAddress:          *e2tAddress,
// 		E2tPort:             *e2tPort,
// 		TopoAddress:         *topoAddress,
// 		TopoPort:            *topoPort,
// 		GRPCPort:            *grpcPort,
// 		RicActionID:         int32(*ricActionID),
// 		SMNameMHO:           *smNamemho,
// 		SMVersionMHO:        *smVersionmho,
// 		SMNameRC:            *smNamerc,
// 		SMVersionRC:         *smVersionrc,
// 		// ConfigPath:          *configPath,
// 		// E2tEndpoint:         *e2tEndpoint,
// 		// UENIBEndpoint:       *uenibEndpoint,
// 		OverloadThreshold:   *overloadThreshold,
// 		TargetLoadThreshold: *targetLoadThreshold,
// 	}

// 	// Initializing Unified Manager with consolidated configurations
// 	unifiedMgr := manager.NewManager(cfg)
// 	unifiedMgr.Run()

// 	log.Info("Unified CCO and MLB Manager has started")
// }
