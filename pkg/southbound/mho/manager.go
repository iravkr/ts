// SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2019-present Rimedo Labs
//
// SPDX-License-Identifier: Apache-2.0
// Created by RIMEDO-Labs team
// based on onosproject/onos-mho/pkg/southbound/e2/manager.go

package mho

import (
	"context"
	"fmt"

	"github.com/onosproject/cco-mon/pkg/mho"
	monitoring "github.com/onosproject/cco-mon/pkg/monitoring/mho"
	"github.com/onosproject/cco-mon/pkg/rnib"

	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_mho_go/pdubuilder"
	e2sm_mho "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_mho_go/v2/e2sm-mho-go"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mho/pkg/broker"
	"github.com/onosproject/onos-mho/pkg/store"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger("cco-mon", "e2", "manager")

const (
	oid = "1.3.6.1.4.1.53148.1.2.2.101"
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

func NewManager(options Options, indCh chan *mho.E2NodeIndication, ctrlReqChs map[string]chan *e2api.ControlMessage, ueStore store.Store, cellStore store.Store) (Manager, error) {

	smName := e2client.ServiceModelName(options.SMName)
	smVer := e2client.ServiceModelVersion(options.SMVersion)
	appID := e2client.AppID(options.AppID)
	e2Client := e2client.NewClient(
		e2client.WithAppID(appID),
		e2client.WithServiceModel(smName, smVer),
		e2client.WithE2TAddress(options.E2tAddress, options.E2tPort),
	)

	rnibOptions := rnib.Options{
		TopoAddress: options.TopoAddress,
		TopoPort:    options.TopoPort,
	}

	rnibClient, err := rnib.NewClient(rnibOptions)
	if err != nil {
		return Manager{}, err
	}

	return Manager{
		e2Client:    e2Client,
		rnibClient:  rnibClient,
		streams:     broker.NewBroker(),
		indCh:       indCh,
		ctrlReqChs:  ctrlReqChs,
		smModelName: smName,
		ueStore:     ueStore,
		cellStore:   cellStore,
	}, nil
}

type Manager struct {
	e2Client    e2client.Client
	rnibClient  rnib.Client
	streams     broker.Broker
	indCh       chan *mho.E2NodeIndication
	ctrlReqChs  map[string]chan *e2api.ControlMessage
	smModelName e2client.ServiceModelName
	ueStore     store.Store
	cellStore   store.Store
}

// start e2 manager instance
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

func (m *Manager) watchE2Connections(ctx context.Context) error {
	ch := make(chan topoapi.Event)
	err := m.rnibClient.WatchE2Connections(ctx, ch)
	if err != nil {
		log.Warn(err)
		return err
	}

	for topoEvent := range ch {
		if topoEvent.Type == topoapi.EventType_ADDED || topoEvent.Type == topoapi.EventType_NONE {
			relation := topoEvent.Object.Obj.(*topoapi.Object_Relation)
			e2NodeID := relation.Relation.TgtEntityID

			triggers := make(map[e2sm_mho.MhoTriggerType]bool)
			triggers[e2sm_mho.MhoTriggerType_MHO_TRIGGER_TYPE_PERIODIC] = true
			triggers[e2sm_mho.MhoTriggerType_MHO_TRIGGER_TYPE_UPON_RCV_MEAS_REPORT] = false
			triggers[e2sm_mho.MhoTriggerType_MHO_TRIGGER_TYPE_UPON_CHANGE_RRC_STATUS] = false

			for triggerType, enabled := range triggers {
				if enabled {
					go func(triggerType e2sm_mho.MhoTriggerType) {
						_ = m.createSubscription(ctx, e2NodeID, triggerType)
					}(triggerType)
				}
			}

		}
	}

	return nil
}

func (m *Manager) createSubscription(ctx context.Context, e2nodeID topoapi.ID, triggerType e2sm_mho.MhoTriggerType) error {
	eventTriggerData, err := m.createEventTrigger(triggerType)
	if err != nil {
		return err
	}
	actions := m.createSubscriptionActions()

	// TO DO: Need to add code to verify that E2 nodes has MHO Service Model //

	ch := make(chan e2api.Indication)
	node := m.e2Client.Node(e2client.NodeID(e2nodeID))
	subName := fmt.Sprintf("cco-mon-subscription-%s", triggerType)
	subSpec := e2api.SubscriptionSpec{
		Actions: actions,
		EventTrigger: e2api.EventTrigger{
			Payload: eventTriggerData,
		},
	}

	channelID, err := node.Subscribe(ctx, subName, subSpec, ch)
	if err != nil {
		log.Warn(err)
		return err
	}

	streamReader, err := m.streams.OpenReader(ctx, node, subName, channelID, subSpec)
	if err != nil {
		return err
	}
	go m.sendIndicationOnStream(streamReader.StreamID(), ch)

	monitor := monitoring.NewMonitor(streamReader, e2nodeID, m.indCh, triggerType)

	err = monitor.Start(ctx)
	if err != nil {
		log.Warn(err)
	}

	return nil
}

func (m *Manager) createEventTrigger(triggerType e2sm_mho.MhoTriggerType) ([]byte, error) {
	var reportPeriodMs int32
	reportingPeriod := 1000

	if triggerType == e2sm_mho.MhoTriggerType_MHO_TRIGGER_TYPE_PERIODIC {
		reportPeriodMs = int32(reportingPeriod)
	} else {
		reportPeriodMs = 0
	}
	e2smRcEventTriggerDefinition, err := pdubuilder.CreateE2SmMhoEventTriggerDefinition(triggerType)
	if err != nil {
		return []byte{}, err
	}
	e2smRcEventTriggerDefinition.GetEventDefinitionFormats().GetEventDefinitionFormat1().SetReportingPeriodInMs(reportPeriodMs)

	err = e2smRcEventTriggerDefinition.Validate()
	if err != nil {
		return []byte{}, err
	}

	protoBytes, err := proto.Marshal(e2smRcEventTriggerDefinition)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, err
}

func (m *Manager) createSubscriptionActions() []e2api.Action {
	actions := make([]e2api.Action, 0)
	action := &e2api.Action{
		ID:   int32(0),
		Type: e2api.ActionType_ACTION_TYPE_REPORT,
		SubsequentAction: &e2api.SubsequentAction{
			Type:       e2api.SubsequentActionType_SUBSEQUENT_ACTION_TYPE_CONTINUE,
			TimeToWait: e2api.TimeToWait_TIME_TO_WAIT_ZERO,
		},
	}
	actions = append(actions, *action)
	return actions
}

func (m *Manager) sendIndicationOnStream(streamID broker.StreamID, ch chan e2api.Indication) {
	streamWriter, err := m.streams.GetWriter(streamID)
	if err != nil {
		log.Error(err)
		return
	}

	for msg := range ch {
		err := streamWriter.Send(msg)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

func (m *Manager) GetCellTypes(ctx context.Context) map[string]rnib.Cell {
	cellTypes, err := m.rnibClient.GetCellTypes(ctx)
	if err != nil {
		log.Warn(err)
	}
	return cellTypes
}
