package northbound

import (
	"fmt"
	"net"
	"sync"

	pb "github.com/onosproject/cco-mon/pkg/proto"
	"github.com/onosproject/cco-mon/pkg/store/metrics"
	ocnstorage "github.com/onosproject/cco-mon/pkg/store/store/ocn"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mho/pkg/store"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("cco-mon", "northbound", "manager")
var server *MonitoringServer

type MonitoringServer struct {
	pb.UnimplementedCCOMonitoringServiceServer
	manager *Manager
}

type Manager struct {
	ueStore     store.Store
	metricStore metrics.Store
	grpcPort    int
	mutex       sync.RWMutex
	ocnStore    ocnstorage.Store
}

func NewManager(grpcPort int, ueStore store.Store, metricStore metrics.Store, ocnStore ocnstorage.Store) *Manager {
	manager := &Manager{
		ueStore:     ueStore,
		metricStore: metricStore,
		grpcPort:    grpcPort,
		mutex:       sync.RWMutex{},
		ocnStore:    ocnStore,
	}

	server = &MonitoringServer{manager: manager}

	return manager
}

func (m *Manager) Start() error {

	//listen on the port
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", m.grpcPort))
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
		return err
	}

	// create a new gRPC server
	grpcServer := grpc.NewServer()

	// register the monitoring service
	pb.RegisterCCOMonitoringServiceServer(grpcServer, server)

	err = grpcServer.Serve(lis)
	log.Infof("CCO Monitoring service is now registered at %v", lis.Addr())

	if err != nil {
		fmt.Printf("Failed to start: %v", err)
		return err
	} else {
		log.Infof("Server started at the following address %v", lis.Addr())

	}

	return nil
}
