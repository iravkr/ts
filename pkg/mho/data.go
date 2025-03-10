// SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>

// SPDX-License-Identifier: Apache-2.0

package mho

import (
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_mho_go/v2/e2sm-v2-ies"
)

type UeData struct {
	UeID          string
	E2NodeID      string
	CGI           *e2sm_v2_ies.Cgi
	CGIString     string
	RrcState      string
	FiveQi        int64
	RsrpServing   int32
	RsrpNeighbors map[string]int32
	RsrpTable     map[string]int32
	CgiTable      map[string]*e2sm_v2_ies.Cgi
	Idle          bool
}

type CellData struct {
	CGI                    *e2sm_v2_ies.Cgi
	CGIString              string
	CumulativeHandoversIn  int
	CumulativeHandoversOut int
	Ues                    map[string]*UeData
}
