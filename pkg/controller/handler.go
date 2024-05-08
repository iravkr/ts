// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"
	"crypto/rand"
	"math/big"
	"fmt"


	"github.com/onosproject/cco-mon/pkg/southbound/e2policy"

	"github.com/onosproject/cco-mon/pkg/monitor"
	ocnstorage "github.com/onosproject/cco-mon/pkg/store/store/ocn"
	paramstorage "github.com/onosproject/cco-mon/pkg/store/store/parameters"
	"github.com/onosproject/cco-mon/pkg/store/store/storage"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	meastype "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
)

var log = logging.GetLogger("Controller", "controller")

const (
	// RcPreRanParamDefaultOCN is default Ocn value
	RcPreRanParamDefaultOCN = meastype.QOffset1dB
)

// NewHandler generates new MLB controller handler
func NewHandler(e2policyHandler e2policy.Handler,
	monitorHandler monitor.Handler,
	numUEsMeasStore storage.Store,
	neighborMeasStore storage.Store,
	ocnStore ocnstorage.Store,
	paramStore paramstorage.Store) Handler {
	return &handler{
		e2PolicyHandler:   e2policyHandler,
		monitorHandler:    monitorHandler,
		numUEsMeasStore:   numUEsMeasStore,
		neighborMeasStore: neighborMeasStore,
		ocnStore:          ocnStore,
		paramStore:        paramStore,
	}
}

// Handler is an interface including MLB controller
type Handler interface {
	// Run runs MLB controller
	Run(ctx context.Context) error
}

type handler struct {
	e2PolicyHandler   e2policy.Handler
	monitorHandler    monitor.Handler
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

func (h *handler) Run(ctx context.Context) error {
	log.Info("Starting the controller")
	for {
		_, err := h.paramStore.Get(context.Background(), "interval")
		if err != nil {
			log.Error(err)
			continue
		}

		select {
		// TODO: swicth to interval
		case <-time.After(time.Duration(1) * time.Second):
			// ToDo should run as goroutine
			log.Info("#handler.run.76 Controller logic before")

			h.startControlLogic(ctx)
		case <-ctx.Done():
			return nil
		}
	}

}


func (h *handler) startControlLogic(ctx context.Context) {
	// run monitor handler
	err := h.monitorHandler.Monitor(ctx)
	log.Infof("interval ")

	if err != nil {
		if err.Error() == monitor.WarnMsgRNIBEmpty {
			log.Warnf(err.Error())
			return
		}
		log.Error(err)
		return
	}

	// update ocn store - to update neighbor or to add new cells coming
	err = h.updateOcnStore(ctx)
	log.Infof("Update Ocn Store test pass")
	log.Infof("Ocn store is %v", h.ocnStore)
	if err != nil {
		log.Error(err)
		return
	}

	// Get total num UE
	totalNumUEs, err := h.getTotalNumUEs(ctx)
	log.Infof("Total number of UEs: %v", totalNumUEs)
	if err != nil {
		log.Error(err)
		return
	}

	// Get Cell IDs
	cells, err := h.getCellList(ctx)
	log.Infof("Cell IDs: %v", cells)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("cells %+v", cells)


	// run control logic for each cell
	for _, cell := range cells {
		log.Infof("Serving cell: %v", cell)
		err = h.controlLogicEachCell_new(ctx, cell, cells, totalNumUEs)
		//err = h.controlLogicEachCell(ctx, cell, cells, totalNumUEs)
		if err != nil {
			// log.Infof("Problem with cell: %v", cell)
			log.Error(err)
			return
		}
	}
}

//What does this update function do. How to translate to txpower?

func (h *handler) updateOcnStore(ctx context.Context) error {
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		log.Info("#handler.updateocnstore.144 before")
		err := h.neighborMeasStore.ListElements(ctx, ch)
		log.Info("#handler.updateocnstore.144 after")
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)

	for e := range ch {
		ids := e.Key
		neighborList := e.Value.([]storage.IDs)

		if _, err := h.ocnStore.Get(ctx, ids); err != nil {
			// the new cells connected
			log.Info("#handlerupdateocnstore neighborList")
			_, err = h.ocnStore.Put(ctx, ids, &ocnstorage.OcnMap{
				Value: make(map[storage.IDs]meastype.QOffsetRange),
			})
			log.Info("#handlerupdateocnstore ocn store hydration")
			if err != nil {
				close(ch)
				return err
			}
			for _, nIDs := range neighborList {
				log.Infof("Adding new neighbor: %v", nIDs)
				err = h.ocnStore.PutInnerMapElem(ctx, ids, nIDs, RcPreRanParamDefaultOCN)
				if err != nil {
					close(ch)
					return err
				}
			}
		} else {
			// delete removed neighbor
			inCh := make(chan ocnstorage.InnerEntry)
			go func(inCh chan ocnstorage.InnerEntry) {
				err := h.ocnStore.ListInnerElement(ctx, ids, inCh)
				if err != nil {
					log.Error(err)
					close(ch)
					close(inCh)
				}
			}(inCh)

			for k := range inCh {
				if !h.containsIDs(k.Key, neighborList) {
					err = h.ocnStore.DeleteInnerElement(ctx, ids, k.Key)
					close(ch)
					return err
				}
			}

			// add new neighbor
			for _, n := range neighborList {
				if _, err = h.ocnStore.GetInnerMapElem(ctx, ids, n); err != nil {
					err = h.ocnStore.PutInnerMapElem(ctx, ids, n, RcPreRanParamDefaultOCN)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (h *handler) containsIDs(ids storage.IDs, idsList []storage.IDs) bool {
	for _, e := range idsList {
		if e == ids {
			return true
		}
	}
	return false
}

func (h *handler) getTotalNumUEs(ctx context.Context) (int, error) {
	result := 0
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		err := h.numUEsMeasStore.ListElements(ctx, ch)
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)
	for e := range ch {
		result += e.Value.(storage.Measurement).Value
	}
	return result, nil
}

func (h *handler) getCellList(ctx context.Context) ([]storage.IDs, error) {
	result := make([]storage.IDs, 0)
	ch := make(chan storage.IDs)
	go func(chan storage.IDs) {
		err := h.numUEsMeasStore.ListKeys(ctx, ch)
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)
	for k := range ch {
		result = append(result, k)
	}
	return result, nil
}

func (h *handler) getCapacity(denominationFactor float64, totalNumUEs int, numUEs int) int {
	capacity := (1 - float64(numUEs)/(denominationFactor*float64(totalNumUEs))) * 100
	return int(capacity)
}

func (h *handler) numUE(ctx context.Context, plmnID string, cid string, cells []storage.IDs) (int, error) {
	storageID, err := h.findIDWithCGI(plmnID, cid, cells)
	if err != nil {
		return 0, err
	}

	entry, err := h.numUEsMeasStore.Get(ctx, storageID)
	if err != nil {
		return 0, err
	}
	return entry.Value.(storage.Measurement).Value, nil
}

func (h *handler) findIDWithCGI(plmnid string, cid string, cells []storage.IDs) (storage.IDs, error) {
	for _, cell := range cells {
		if cell.PlmnID == plmnid && cell.CellID == cid {
			return cell, nil
		}
	}
	return storage.IDs{}, errors.NewNotFound("ID not found with plmnid and cgi")
}
func (h *handler) controlLogicEachCell_new(ctx context.Context, ids storage.IDs, cells []storage.IDs, totalNumUEs int) error {

	targetThreshold, err := h.paramStore.Get(context.Background(), "target_threshold")
	if err != nil {
		log.Infof("Problem 1: %v", err)
		return err
	}
	overloadThreshold, err := h.paramStore.Get(context.Background(), "overload_threshold")
	if err != nil {
		log.Infof("Problem 2: %v", err)
		return err
	}
	/*
		ocnDeltaFactor, err := h.paramStore.Get(context.Background(), "delta_ocn")
		if err != nil {
			log.Infof("Problem 3: %v", err)
			return err
		}
	*/
	neighbors, err := h.neighborMeasStore.Get(ctx, ids)
	if err != nil {
		log.Infof("Problem 4 neigh store: %v", err)
		return err
	}

	// calculate for each capacity and check sCell's and its neighbors' capacity
	// if sCell load < target load threshold
	// reduce Ocn
	neighborList := neighbors.Value.([]storage.IDs)
	numUEsSCell, err := h.numUE(ctx, ids.PlmnID, ids.CellID, cells)
	if err != nil {
		return err
	}

	capSCell := h.getCapacity(1, totalNumUEs, numUEsSCell)
	log.Infof("Serving cell (%v) capacity: %v, load: %v / neighbor: %v / overload threshold %v, target threshold %v", ids, capSCell, 100-capSCell, cells, overloadThreshold, targetThreshold)

	// if sCell load > overload threshold && nCell < target load threshold (Overload Control Notification) = CIO
	// increase Ocn

	tmpOcns := make(map[storage.IDs]meastype.QOffsetRange)
	for _, nCellID := range neighborList {
		ocn, err := h.ocnStore.GetInnerMapElem(ctx, ids, nCellID) //
		log.Info("Neighbor cell (%v) OCN: %v", nCellID, ocn)
		if err != nil {
			return err
		}
		// Random Exploration code

		// rand.New(rand.NewSource(335))

		// Generate a random number between 10 and 24
	// 	n, err := rand.Int(rand.Reader, big.NewInt(15))
	// if err != nil {
	// 	panic(err) // rand.Int should not fail under normal circumstances
	// }

	// // Add 10 to shift the range to [10, 24]
	// randomNumber := n.Int64() + 10
		randomNumber, err := generateRandomNumber(-24, 24)
		if err != nil {
			fmt.Printf("Error generating random number: %v\n", err)
			// return
		}


		log.Infof("Random ocn in cid %v: %v", ids, randomNumber)
		tmpOcns[nCellID] = meastype.QOffsetRange(randomNumber)
		log.Infof("To understand the data structure random value  meastype.QOffsetRange: %v", meastype.QOffsetRange(randomNumber))
		/*log.Infof("To understand the data structure -1th value  meastype.QOffsetRange: %v", meastype.QOffsetRange(-1))
		log.Infof("To understand the data structure 1st value  meastype.QOffsetRange: %v", meastype.QOffsetRange(1))
		log.Infof("To understand the data structure 6th value  meastype.QOffsetRange: %v", meastype.QOffsetRange(6))
		log.Infof("To understand the data structure -6th value  meastype.QOffsetRange: %v", meastype.QOffsetRange(-6))
		//tmpOcns[nCellID] = tmpOcns[nCellID] + meastype.QOffsetRange(ocnDeltaFactor)*/
		log.Infof("Final tmpOcns content: %+v", tmpOcns)
		//h.ocnStore.UpdateInnerMapElem(ctx,ids,nCellID,meastype.QOffsetRange(randomNumber))
		err = h.ocnStore.PutInnerMapElems(ctx, ids, tmpOcns)
		if err != nil {
			log.Errorf("Failed to update OCN values for cell %v: %v", ids, err)
			return err
		}

		// time.Sleep(10 * time.Second)
		// err = h.ocnStore.PutInnerMapElem(ctx, ids, nCellID, RcPreRanParamDefaultOCN)



	}
	time.Sleep(5 * time.Second)

	err = h.e2PolicyHandler.SetPolicyForOcn(ctx, ids.NodeID, tmpOcns)
	if err != nil {
		return err
	}
	err = h.ocnStore.PutInnerMapElems(ctx, ids, tmpOcns)
	if err != nil {
		return err
	}

	return nil
}
func generateRandomNumber(min, max int64) (int64, error) {
	if min > max {
		return 0, fmt.Errorf("invalid range: %d > %d", min, max)
	}
	// Compute the range size (add 1 to include max)
	rangeSize := big.NewInt(max - min + 1)
	// Generate a random number in [0, rangeSize)
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}
	// Shift the range to [min, max]
	return n.Int64() + min, nil
}
