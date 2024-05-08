// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

//Monitor receives the indication message from RC SM, extracts cells info and store the data into metrics store

package monitoring

import (
	"context"

	"github.com/onosproject/cco-mon/pkg/rnib"
	"github.com/onosproject/cco-mon/pkg/store/metrics"
	"github.com/onosproject/cco-mon/pkg/utils/decode"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mho/pkg/broker"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger()

// NewMonitor creates a new indication monitor
func NewMonitor(metricStore metrics.Store, node e2client.Node, streamReader broker.StreamReader, e2nodeID topoapi.ID, rnibClient rnib.Client) *Monitor {
	return &Monitor{
		streamReader: streamReader,
		metricStore:  metricStore,
		nodeID:       e2nodeID,
		rnibClient:   rnibClient,
	}
}

// Monitor indication monitor
type Monitor struct {
	streamReader broker.StreamReader
	metricStore  metrics.Store
	nodeID       topoapi.ID
	rnibClient   rnib.Client
}

func (m *Monitor) processIndicationFormat3(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	header := e2smrc.E2SmRcIndicationHeader{}
	err := proto.Unmarshal(indication.Header, &header)
	if err != nil {
		return err
	}

	message := e2smrc.E2SmRcIndicationMessage{}
	err = proto.Unmarshal(indication.Payload, &message)
	if err != nil {
		return err
	}

	headerFormat1 := header.GetRicIndicationHeaderFormats().GetIndicationHeaderFormat1()
	messageFormat3 := message.GetRicIndicationMessageFormats().GetIndicationMessageFormat3()

	log.Infof("Indication header format 1 %v", headerFormat1)
	log.Infof("Indication message format 3 %v", messageFormat3)

	for _, cellInfo := range messageFormat3.GetCellInfoList() {
		if cellInfo.GetCellGlobalId().GetNRCgi() != nil {
			log.Info("5G case")
			cgi := cellInfo.GetCellGlobalId()
			if cellInfo.GetNeighborRelationTable().GetServingCellPci().GetNR() == nil {
				log.Errorf("PCI should be NR PCI but NR PCI field is empty in E2 Indication message")
				continue
			}
			pci := cellInfo.GetNeighborRelationTable().GetServingCellPci().GetNR().GetValue()
			if cellInfo.GetNeighborRelationTable().GetServingCellArfcn().GetNR() == nil {
				log.Errorf("ARFCN should be NR ARFCN but NR ARFCN field is empty in E2 indication message")
				continue
			}
			key := metrics.NewKey(cgi)
			_, err := m.metricStore.Put(ctx, key, metrics.Entry{
				Key: metrics.Key{
					CellGlobalID: cgi,
				},
				Value: metrics.CellMetric{
					E2NodeID:    nodeID,
					CGIstring:   decode.GetCGIString(cgi),
					PTX:         float64(pci),
					PreviousPTX: float64(pci),
				},
			})
			if err != nil {
				return err
			}
		} else {
			// 4G case
			// ToDo: Add 4G case here
			log.Errorf("4G case is not implemented yet")
		}
	}
	return nil
}

func (m *Monitor) processIndication(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	err := m.processIndicationFormat3(ctx, indication, nodeID)
	if err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

// Start start monitoring of indication messages for a given subscription ID
func (m *Monitor) Start(ctx context.Context) error {
	log.Info("Starting the monitoring agent")
	errCh := make(chan error)
	go func() {
		for {
			indMsg, err := m.streamReader.Recv(ctx)
			if err != nil {
				errCh <- err
			}
			err = m.processIndication(ctx, indMsg, m.nodeID)
			if err != nil {
				errCh <- err
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
