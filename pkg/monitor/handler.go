// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"context"
	"fmt"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/cco-mon/pkg/rnib"
	ocnstorage "github.com/onosproject/cco-mon/pkg/store/store/ocn"
	"github.com/onosproject/cco-mon/pkg/store/store/storage"
)

var log = logging.GetLogger()

const (
	WarnMsgRNIBEmpty = "R-NIB does not have enough information - either KPIMON monitoring result or neighbor information is missing"
)

// NewHandler generates monitoring handler
func NewHandler(rnibHandler rnib.Handler, numUEsMeasStore storage.Store, neighborMeasStore storage.Store, ocnStore ocnstorage.Store) Handler {
	return &handler{
		rnibHandler:       rnibHandler,
		numUEsMeasStore:   numUEsMeasStore,
		neighborMeasStore: neighborMeasStore,
		ocnStore:          ocnStore,
	}
}

// Handler is an interface including this handler's functions
type Handler interface {
	// Monitor starts to monitor UENIB and RNIB
	Monitor(ctx context.Context) error
}

type handler struct {
	rnibHandler       rnib.Handler
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
}

func (h *handler) Monitor(ctx context.Context) error {
	// get RNIB
	rnibList, err := h.rnibHandler.Get(ctx)
	if err != nil {
		return err
	} else if len(rnibList) == 0 {
		return fmt.Errorf(WarnMsgRNIBEmpty)
	}

	// store monitoring result
	h.storeRNIB(ctx, rnibList)

	log.Infof("RNIB List %v", rnibList)

	return nil
}

func (h *handler) storeRNIB(ctx context.Context, rnibList []rnib.Element) {
	log.Info("Store RNIB Handler")
	log.Infof("RNIB List %v", rnibList)
	for _, e := range rnibList {
		key := storage.IDs{
			NodeID:    e.Key.IDs.E2NodeID,
			PlmnID:    e.Key.IDs.CellGlobalID.PlmnID,
			CellID:    e.Key.IDs.CellGlobalID.CellIdentity,
			CellObjID: e.Key.IDs.CellObjectID,
		}
		switch e.Key.Aspect {
		case rnib.Neighbors:
			err := h.storeRNIBNeighbors(ctx, key, e.Value.([]rnib.CellGlobalID))
			if err != nil {
				log.Error(err)
			}
		case rnib.NumUEs:
			err := h.storeRNIBNumUEs(ctx, key, e.Value.(uint32))
			if err != nil {
				log.Error(err)
			}
		default:
			log.Warnf("Unavailable aspects for this app - to be discarded: %v", e.Key.Aspect.String())
		}
	}
}

func (h *handler) storeRNIBNeighbors(ctx context.Context, key storage.IDs, neighborIDs []rnib.CellGlobalID) error {
	//log.Info("Store RNIB Neighbors Handler")
	//log.Infof("RNIB Neighbors List %v", neighborIDs)
	nidList := make([]storage.IDs, 0)
	for _, id := range neighborIDs {
		nid := storage.IDs{
			PlmnID: id.PlmnID,
			CellID: id.CellIdentity,
		}
		nidList = append(nidList, nid)
	}
	_, err := h.neighborMeasStore.Put(ctx, key, nidList)
	return err
}

func (h *handler) storeRNIBNumUEs(ctx context.Context, key storage.IDs, value uint32) error {
	measurement := storage.Measurement{
		Value: int(value),
	}
	_, err := h.numUEsMeasStore.Put(ctx, key, measurement)
	return err
}
