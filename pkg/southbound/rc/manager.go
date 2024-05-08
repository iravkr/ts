// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package rc

import (
	"context"

	monitoring "github.com/onosproject/cco-mon/pkg/monitoring/rc"
	"github.com/onosproject/cco-mon/pkg/rnib"
	"github.com/onosproject/cco-mon/pkg/store/metrics"
	"github.com/onosproject/cco-mon/pkg/utils/control"
	subutils "github.com/onosproject/cco-mon/pkg/utils/subscription"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mho/pkg/broker"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
)

var log = logging.GetLogger()

const (
	oid = "1.3.6.1.4.1.53148.1.1.2.3"
)

type Options struct {
	AppID       string
	E2tAddress  string
	E2tPort     int
	TopoAddress string
	TopoPort    int
	SMName      string
	SMVersion   string
}

// Node e2 manager interface
type Node interface {
	Start() error
	Stop() error
}

// ServiceModelOptions is options for defining a service model
type ServiceModelOptions struct {
	// Name is the service model identifier
	Name string

	// Version is the service model version
	Version string
}

// Manager subscription manager
type Manager struct {
	e2client     e2client.Client
	rnibClient   rnib.Client
	serviceModel ServiceModelOptions
	streams      broker.Broker
	metricStore  metrics.Store
}

// NewManager creates a new subscription manager
func NewManager(opts Options, metricStore metrics.Store) (Manager, error) {

	serviceModelName := e2client.ServiceModelName(opts.SMName)
	serviceModelVersion := e2client.ServiceModelVersion(opts.SMVersion)
	appID := e2client.AppID(opts.AppID)
	e2Client := e2client.NewClient(
		e2client.WithServiceModel(serviceModelName, serviceModelVersion),
		e2client.WithAppID(appID),
		e2client.WithE2TAddress(opts.E2tAddress, opts.E2tPort))

	rnibOptions := rnib.Options{
		TopoAddress: opts.TopoAddress,
		TopoPort:    opts.TopoPort,
	}

	rnibClient, err := rnib.NewClient(rnibOptions)
	if err != nil {
		return Manager{}, err
	}

	return Manager{
		e2client:   e2Client,
		rnibClient: rnibClient,
		serviceModel: ServiceModelOptions{
			Name:    opts.SMName,
			Version: opts.SMVersion,
		},
		streams:     broker.NewBroker(),
		metricStore: metricStore,
	}, nil

}

// Start starts subscription manager
func (m *Manager) Start() error {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := m.watchE2Connections(ctx)
		if err != nil {
			return
		}
	}()

	return nil
}

func (m *Manager) sendIndicationOnStream(streamID broker.StreamID, ch chan e2api.Indication) {
	log.Info("Start sending indication on stream")
	streamWriter, err := m.streams.GetWriter(streamID)
	log.Infof("Stream Writer %v", streamWriter)
	if err != nil {
		return
	}

	for msg := range ch {
		err := streamWriter.Send(msg)
		log.Infof("Msg sent %v", msg)
		if err != nil {
			log.Info(err)
			log.Warn(err)
			return
		}
	}
}

/*func (m *Manager) getRanFunction(serviceModelsInfo map[string]*topoapi.ServiceModelInfo) (*topoapi.RCRanFunction, error) {
	log.Info("Service models:", serviceModelsInfo)
	for _, sm := range serviceModelsInfo {
		smName := strings.ToLower(sm.Name)
		if smName == string(m.serviceModel.Name) && sm.OID == oid {
			rcRanFunction := &topoapi.RCRanFunction{}
			for _, ranFunction := range sm.RanFunctions {
				if ranFunction.TypeUrl == ranFunction.GetTypeUrl() {
					err := prototypes.UnmarshalAny(ranFunction, rcRanFunction)
					if err != nil {
						return nil, err
					}
					return rcRanFunction, nil
				}
			}
		}
	}
	return nil, errors.New(errors.NotFound, "cannot retrieve ran functions")

}*/

func (m *Manager) createSubscription(ctx context.Context, e2nodeID topoapi.ID) error {
	log.Infof("Creating subscription for E2 node with ID: ", e2nodeID)
	eventTriggerData, err := subutils.CreateEventTriggerDefinition()
	if err != nil {
		log.Warn(err)
		return err
	}
	/*aspects, err := m.rnibClient.GetE2NodeAspects(ctx, e2nodeID)
	if err != nil {
		log.Warn(err)
		return err
	}*/

	/*_, err = m.getRanFunction(aspects.ServiceModels)
	if err != nil {
		log.Warn(err)
		return err
	}*/

	actions := subutils.CreateSubscriptionActions()
	log.Infof("subscription actions %v", actions)

	ch := make(chan e2api.Indication)
	node := m.e2client.Node(e2client.NodeID(e2nodeID))
	log.Infof("Node %v", node)
	subName := "onos-pci-subscription"
	subSpec := e2api.SubscriptionSpec{
		Actions: actions,
		EventTrigger: e2api.EventTrigger{
			Payload: eventTriggerData,
		},
	}
	channelID, err := node.Subscribe(ctx, subName, subSpec, ch)
	log.Infof("Channel ID %v", channelID)
	if err != nil {
		log.Warn(err)
		return err
	}
	log.Debugf("Channel ID:%s", channelID)
	streamReader, err := m.streams.OpenReader(ctx, node, subName, channelID, subSpec)
	log.Infof("Stream reader %v", streamReader)
	if err != nil {
		return err
	}

	go m.sendIndicationOnStream(streamReader.StreamID(), ch)
	monitor := monitoring.NewMonitor(m.metricStore, node, streamReader, e2nodeID, m.rnibClient)

	log.Infof("Monitor %v", monitor)
	err = monitor.Start(ctx)
	log.Infof("Monitor %v started with %v", monitor, err)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil

}

func (m *Manager) newSubscription(ctx context.Context, e2NodeID topoapi.ID) error {
	err := m.createSubscription(ctx, e2NodeID)
	return err
}

func (m *Manager) watchE2Connections(ctx context.Context) error {
	ch := make(chan topoapi.Event)
	err := m.rnibClient.WatchE2Connections(ctx, ch)
	if err != nil {
		log.Warn(err)
		return err
	}

	// creates a new subscription whenever there is a new E2 node connected and supports KPM service model
	for topoEvent := range ch {
		if topoEvent.Type == topoapi.EventType_ADDED || topoEvent.Type == topoapi.EventType_NONE {
			log.Infof("New E2 connection detected")
			relation := topoEvent.Object.Obj.(*topoapi.Object_Relation)
			e2NodeID := relation.Relation.TgtEntityID

			if !m.rnibClient.HasRCRANFunction(ctx, e2NodeID, oid) {
				log.Debugf("Received topo event does not have RC RAN function - %v", topoEvent)
				continue
			}

			go func() {
				log.Debugf("start creating subscriptions %v", topoEvent)
				err := m.newSubscription(ctx, e2NodeID)
				if err != nil {
					log.Warn(err)
				}
			}()
			go m.watchPTXChanges(ctx, e2NodeID)
		}

	}
	return nil
}

func (m *Manager) watchPTXChanges(ctx context.Context, e2nodeID topoapi.ID) {
	log.Info("Watching PTX changes")
	ch := make(chan metrics.Event)
	err := m.metricStore.Watch(ctx, ch)
	if err != nil {
		return
	}

	for e := range ch {
		log.Infof("received some data, e.type %v (%v), & e2nodeID %v", e.Type, metrics.UpdatedPTX, e.Value.Value.E2NodeID)
		if e.Type == metrics.UpdatedPTX && e2nodeID == e.Value.Value.E2NodeID {
			key := e.Value.Key
			header, err := control.CreateRcControlHeader(key.CellGlobalID)
			if err != nil {
				log.Warn(err)
			}
			newPtx := e.Value.Value.PTX
			log.Infof("send control message for key: %v, old ptx %v, new ptx: %v", key.CellGlobalID, e.Value.Value.PreviousPTX, newPtx)
			payload, err := control.CreateRcControlMessage(int64(newPtx), key.CellGlobalID)
			if err != nil {
				log.Warn(err)
			}

			node := m.e2client.Node(e2client.NodeID(e2nodeID))
			outcome, err := node.Control(ctx, &e2api.ControlMessage{
				Header:  header,
				Payload: payload,
			})

			if err != nil {
				log.Warn(err)
			}
			log.Infof("Outcome:%v", outcome)
		}
	}

}

// Stop stops the subscription manager
func (m *Manager) Stop() error {
	panic("implement me")
}

// GetMetricsStore returns the metrics store
func (m *Manager) GetMetricsStore() metrics.Store {
	return m.metricStore
}

var _ Node = &Manager{}
