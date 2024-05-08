// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
)

// Key metric key
type Key struct {
	CellGlobalID *e2smrccomm.Cgi
}

// Entry entry of metrics store
type Entry struct {
	Key   Key
	Value CellMetric
}

type CellMetric struct {
	E2NodeID    topoapi.ID
	CGIstring   string
	PTX         float64
	PreviousPTX float64
}

const (
	LowerPCI = 1
	UpperPCI = 503
)

// MetricEvent a metric event
type MetricEvent int

const (
	// None none cell event
	None MetricEvent = iota
	// Created created measurement event
	Created
	// Updated updated measurement event
	Updated
	// UpdatedPTX updated PTX in measurement
	UpdatedPTX
	// Deleted deleted measurement event
	Deleted
)

func (e MetricEvent) String() string {
	return [...]string{"None", "Created", "Updated", "UpdatedPTX", "Deleted"}[e]
}
