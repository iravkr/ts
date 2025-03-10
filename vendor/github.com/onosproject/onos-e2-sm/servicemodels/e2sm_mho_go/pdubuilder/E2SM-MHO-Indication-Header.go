// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package pdubuilder

import (
	"fmt"
	e2sm_mho_go "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_mho_go/v2/e2sm-mho-go"
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_mho_go/v2/e2sm-v2-ies"
)

func CreateE2SmMhoIndicationHeader(cgi *e2sm_v2_ies.Cgi) (*e2sm_mho_go.E2SmMhoIndicationHeader, error) {

	E2SmMhoPdu := e2sm_mho_go.E2SmMhoIndicationHeader{
		E2SmMhoIndicationHeader: &e2sm_mho_go.E2SmMhoIndicationHeader_IndicationHeaderFormat1{
			IndicationHeaderFormat1: &e2sm_mho_go.E2SmMhoIndicationHeaderFormat1{
				Cgi: cgi,
			},
		},
	}

	if err := E2SmMhoPdu.Validate(); err != nil {
		return nil, fmt.Errorf("CreateE2SmMhoIndicationHeader(): error validating E2SmMhoPDU %s", err.Error())
	}
	return &E2SmMhoPdu, nil
}
