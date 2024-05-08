// SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2019-present Rimedo Labs
//
// SPDX-License-Identifier: Apache-2.0
// Created by RIMEDO-Labs team

// rnib client retrieves cells info from the topology componenet (network information base)

package rnib

import (
	"context"
	"fmt"

	idutils "github.com/onosproject/cco-mon/pkg/utils/parse"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-ric-sdk-go/pkg/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

var log = logging.GetLogger("cco-mon", "rnib")

const (
	// AspectKeyNumUEsRANSim is the R-NIB aspect key of the number of UEs for RAN-Simulator
	AspectKeyNumUEsRANSim = "RRC.Conn.Avg"

	// AspectKeyNumUEsOAI is the R-NIB aspect key of the number of UEs for OAI
	AspectKeyNumUEsOAI = "RRC.ConnMean"
)

type TopoClient interface {
	WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error
}

type Options struct {
	TopoAddress string
	TopoPort    int
}

type Cell struct {
	CGI      string
	CellType string
}

func NewClient(options Options) (Client, error) {
	sdkClient, err := toposdk.NewClient(
		toposdk.WithTopoAddress(
			options.TopoAddress,
			options.TopoPort,
		),
	)
	if err != nil {
		return Client{}, err
	}
	return Client{
		client: sdkClient,
	}, nil
}

// NewHandler generates the new RNIB handler
func NewHandler(options Options) (Handler, error) {
	rnibClient, err := topo.NewClient(toposdk.WithTopoAddress(
		options.TopoAddress,
		options.TopoPort,
	))
	if err != nil {
		return nil, err
	}
	return &handler{
		rnibClient: rnibClient,
	}, nil
}

type Client struct {
	client toposdk.Client
}

func getControlRelationFilter() *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.CONTROLS,
				},
			},
		},
	}
	return controlRelationFilter
}

func (c *Client) WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error {
	err := c.client.Watch(ctx, ch, toposdk.WithWatchFilters(getControlRelationFilter()))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetE2CellFilter() *topoapi.Filters {
	cellEntityFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_In{
				In: &topoapi.InFilter{
					Values: []string{topoapi.E2CELL},
				},
			},
		},
	}
	return cellEntityFilter
}

func (c *Client) GetCellTypes(ctx context.Context) (map[string]Cell, error) {
	output := make(map[string]Cell)

	cells, err := c.client.List(ctx, toposdk.WithListFilters(c.GetE2CellFilter()))
	if err != nil {
		log.Warn(err)
		return output, err
	}

	for _, cell := range cells {

		cellObject := &topoapi.E2Cell{}
		err = cell.GetAspect(cellObject)
		if err != nil {
			log.Warn(err)
		}
		output[string(cell.ID)] = Cell{
			CGI:      cellObject.CellObjectID,
			CellType: cellObject.CellType,
		}
	}
	return output, nil
}

func (c *Client) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := c.client.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	e2Node := &topoapi.E2Node{}
	err = object.GetAspect(e2Node)

	return e2Node, err

}

func (c *Client) HasRCRANFunction(ctx context.Context, nodeID topoapi.ID, oid string) bool {
	e2Node, err := c.GetE2NodeAspects(ctx, nodeID)
	if err != nil {
		log.Warn(err)
		return false
	}

	for _, sm := range e2Node.GetServiceModels() {
		if sm.OID == oid {
			return true
		}
	}
	return false
}

// GetCells get list of cells for each E2 node
func (c *Client) GetCells(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.E2Cell, error) {
	filter := &topoapi.Filters{
		RelationFilter: &topoapi.RelationFilter{SrcId: string(nodeID),
			RelationKind: topoapi.CONTAINS,
			TargetKind:   ""}}

	objects, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	if err != nil {
		return nil, err
	}
	var cells []*topoapi.E2Cell
	for _, obj := range objects {
		targetEntity := obj.GetEntity()
		if targetEntity.GetKindID() == topoapi.E2CELL {
			cellObject := &topoapi.E2Cell{}
			err = obj.GetAspect(cellObject)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cellObject)
		}
	}
	return cells, nil
}

// E2NodeIDs lists all of connected E2 nodes
func (c *Client) E2NodeIDs(ctx context.Context, oid string) ([]topoapi.ID, error) {
	objects, err := c.client.List(ctx, toposdk.WithListFilters(getControlRelationFilter()))
	if err != nil {
		return nil, err
	}

	e2NodeIDs := make([]topoapi.ID, len(objects))
	for _, object := range objects {
		relation := object.Obj.(*topoapi.Object_Relation)
		e2NodeID := relation.Relation.TgtEntityID
		if c.HasRCRANFunction(ctx, e2NodeID, oid) {
			e2NodeIDs = append(e2NodeIDs, e2NodeID)
		}
	}

	return e2NodeIDs, nil
}

// Handler includes RNIB handler's all functions
type Handler interface {
	// Get gets all RNIB
	Get(ctx context.Context) ([]Element, error)
	GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error)
}

type handler struct {
	rnibClient topo.Client
}

func (h *handler) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := h.rnibClient.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	e2Node := &topoapi.E2Node{}
	err = object.GetAspect(e2Node)

	return e2Node, err
}

func (h *handler) Get(ctx context.Context) ([]Element, error) {
	objects, err := h.rnibClient.List(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	result := make([]Element, 0)

	// log.Debugf("R-NIB objects - %s", objects)
	for _, obj := range objects {
		if obj.GetEntity() == nil || obj.GetEntity().GetKindID() != topoapi.E2CELL {
			continue
		}
		// log.Debugf("R-NIB each obj: %s", obj)
		cellTopoID := obj.GetID()
		e2NodeID, cellIdentity := idutils.ParseCellTopoID(string(cellTopoID))
		cellObject := topoapi.E2Cell{}
		err = obj.GetAspect(&cellObject)
		if err != nil {
			return nil, err
		}

		cellObjectID := cellObject.CellObjectID
		if cellIdentity != cellObject.CellGlobalID.GetValue() {
			return nil, fmt.Errorf("verification failed: In R-NIB, cell IDs in topo ID field and aspects are different")
		}
		// ToDo: add PLMN ID here for cell object in the future
		plmnID := ""

		if cellObjectID == "" || cellIdentity == "" {
			return nil, fmt.Errorf("R-NIB is not ready yet")
		}

		ids := IDs{
			TopoID:       cellTopoID,
			E2NodeID:     e2NodeID,
			CellObjectID: cellObjectID,
			CellGlobalID: CellGlobalID{
				CellIdentity: cellIdentity,
				PlmnID:       plmnID,
			},
		}

		if len(cellObject.NeighborCellIDs) == 0 || len(cellObject.KpiReports) == 0 {
			continue
		}

		neighbors := make([]CellGlobalID, 0)
		for _, neighborCellID := range cellObject.NeighborCellIDs {
			neighborCellGlobalID := CellGlobalID{
				CellIdentity: neighborCellID.CellGlobalID.GetValue(),
				PlmnID:       neighborCellID.PlmnID,
			}
			neighbors = append(neighbors, neighborCellGlobalID)
			plmnID = neighborCellID.PlmnID
		}
		ids.CellGlobalID.PlmnID = plmnID
		neighborElement := Element{
			Key: Key{
				IDs:    ids,
				Aspect: Neighbors,
			},
			Value: neighbors,
		}
		result = append(result, neighborElement)

		for kpiKey, kpiValue := range cellObject.KpiReports {
			if kpiKey == AspectKeyNumUEsOAI || kpiKey == AspectKeyNumUEsRANSim {
				kpiElement := Element{
					Key: Key{
						IDs:    ids,
						Aspect: NumUEs,
					},
					Value: kpiValue,
				}
				result = append(result, kpiElement)
				break
			}
		}
	}

	return result, nil
}
