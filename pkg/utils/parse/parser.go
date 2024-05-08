// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package parse

import (
	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"fmt"
	"github.com/onosproject/onos-api/go/onos/uenib"
	"strings"

)

type CGIType int

const (
	CGITypeNrCGI CGIType = iota
	CGITypeECGI
	CGITypeUnknown
)

func (c CGIType) String() string {
	return [...]string{"CGITypeNRCGI", "CGITypeECGI", "CGITypeUnknown"}[c]
}

func GetNRMetricKey(e *e2smrccomm.NrCgi) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	}
	return e.GetPLmnidentity().GetValue(),
		BitStringToUint64(e.GetNRcellIdentity().GetValue().GetValue(), int(e.GetNRcellIdentity().GetValue().GetLen())),
		CGITypeNrCGI,
		nil
}

func GetEUTRAMetricKey(e *e2smrccomm.EutraCgi) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	}
	return e.GetPLmnidentity().GetValue(),
		BitStringToUint64(e.GetEUtracellIdentity().GetValue().GetValue(), int(e.GetEUtracellIdentity().GetValue().GetLen())),
		CGITypeECGI,
		nil
}

func GetCellID(cellGlobalID *e2smrccomm.Cgi) (uint64, error) {
	switch v := cellGlobalID.Cgi.(type) {
	case *e2smrccomm.Cgi_EUtraCgi:
		return BitStringToUint64(v.EUtraCgi.GetEUtracellIdentity().GetValue().GetValue(), int(v.EUtraCgi.GetEUtracellIdentity().GetValue().GetLen())), nil
	case *e2smrccomm.Cgi_NRCgi:
		return BitStringToUint64(v.NRCgi.GetNRcellIdentity().GetValue().GetValue(), int(v.NRCgi.GetNRcellIdentity().GetValue().GetLen())), nil
	}
	return 0, errors.New(errors.NotSupported, "CGI should be one of NrCGI and ECGI")
}

func BitStringToUint64(bitString []byte, bitCount int) uint64 {
	var result uint64
	for i, b := range bitString {
		result += uint64(b) << ((len(bitString) - i - 1) * 8)
	}
	if bitCount%8 != 0 {
		return result >> (8 - bitCount%8)
	}
	return result
}
// ParseUENIBNeighborAspectKey parses neighbor aspect key in UENIB
func ParseUENIBNeighborAspectKey(key uenib.ID) (string, string, string, string, error) {
	// ToDo: PCI app should store this with hex format
	objects := strings.Split(string(key), ":")
	if len(objects) != 4 {
		return "", "", "", "", errors.NewNotSupported("neighbor aspect's key should have four key elements")
	}

	nodeID := objects[0]
	plmnID := objects[1]
	cid := objects[2]
	ecgiType := objects[3]

	return nodeID, plmnID, cid, ecgiType, nil
}

// ParseUENIBNeighborAspectValue parses neighbor aspect value in UENIB
func ParseUENIBNeighborAspectValue(value string) (string, error) {
	// ToDo: PCI app should store this with hex format
	results := ""
	idsList := strings.Split(value, ",")
	for _, ids := range idsList {
		idList := strings.Split(ids, ":")
		plmnID := idList[0]
		cid := idList[1]
		if results == "" {
			results = fmt.Sprintf("%s:%s:%s", plmnID, cid, idList[2])
			continue
		}
		results = fmt.Sprintf("%s,%s:%s:%s", results, plmnID, cid, idList[2])
	}
	return results, nil
}

// ParseUENIBNumUEsAspectKey parses the number of UEs aspect key in UENIB
func ParseUENIBNumUEsAspectKey(key uenib.ID) (string, string, error) {
	objects := strings.Split(string(key), ":")
	if len(objects) != 2 {
		return "", "", errors.NewNotSupported("aspect's key for the number of UEs should have two key elements")
	}

	nodeID := objects[0]
	coi := objects[1]
	return nodeID, coi, nil
}

// Uint64ToBitString converts uint64 to a bit string byte array
func Uint64ToBitString(value uint64, bitCount int) []byte {
	result := make([]byte, bitCount/8+1)
	if bitCount%8 > 0 {
		value = value << (8 - bitCount%8)
	}

	for i := 0; i <= (bitCount / 8); i++ {
		result[i] = byte(value >> (((bitCount / 8) - i) * 8) & 0xFF)
	}

	return result
}

// ParseCellTopoID parses TopoID to nodeID and cellID
func ParseCellTopoID(cellTopoID string) (string, string) {
	ids := strings.Split(cellTopoID, "/")
	return fmt.Sprintf("%s/%s", ids[0], ids[1]), ids[2]
}
