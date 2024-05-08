// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package decode

import (
	"fmt"
	"strconv"
	"encoding/binary"
	"bytes"

	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
)

// PlmnIDToUint32 decodes PLMN ID from byte array to uint32
func PlmnIDToUint32(plmnBytes []byte) uint32 {
	return uint32(plmnBytes[0]) | uint32(plmnBytes[1])<<8 | uint32(plmnBytes[2])<<16
}

///

func PlmnIDBytesToInt(b []byte) uint64 {
	return uint64(b[2])<<16 | uint64(b[1])<<8 | uint64(b[0])
}

func PlmnIDNciToCGI(plmnID uint64, nci uint64) string {
	cgi := strconv.FormatInt(int64(plmnID<<36|(nci&0xfffffffff)), 16)
	return cgi
}

func GetNciFromCellGlobalID(cellGlobalID *e2smrccomm.Cgi) uint64 {
	return BitStringToUint64(cellGlobalID.GetNRCgi().GetNRcellIdentity().GetValue().GetValue(), int(cellGlobalID.GetNRCgi().GetNRcellIdentity().GetValue().GetLen()))
}

func GetPlmnIDBytesFromCellGlobalID(cellGlobalID *e2smrccomm.Cgi) []byte {
	return cellGlobalID.GetNRCgi().GetPLmnidentity().GetValue()
}

func GetMccMncFromPlmnID(plmnId uint64) (string, string) {
	plmnIdString := strconv.FormatUint(plmnId, 16)
	return plmnIdString[0:3], plmnIdString[3:]
}

func GetPlmnIdFromMccMnc(mcc string, mnc string) (uint64, error) {
	combined := mcc + mnc
	plmnId, err := strconv.ParseUint(combined, 16, 64)
	if err != nil {
		fmt.Printf("Cannot convert PLMN ID string into uint64 type! %v", err)
	}
	return plmnId, err
}

func GetCGIString(cgi *e2smrccomm.Cgi) string {
	nci := GetNciFromCellGlobalID(cgi)
	plmnIDBytes := GetPlmnIDBytesFromCellGlobalID(cgi)
	plmnID := PlmnIDBytesToInt(plmnIDBytes)
	return PlmnIDNciToCGI(plmnID, nci)
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


// DecodePlmnIDDecStrToBytes decodes the PLMNID dec string type to bytes type
func DecodePlmnIDDecStrToBytes(plmnidDecStr string) ([]byte, error) {
	var plmnBytes [3]uint8
	n, err := strconv.ParseUint(plmnidDecStr, 10, 32)
	if err != nil {
		return nil, err
	}
	plmnid := uint32(n)

	plmnBytes[0] = uint8(plmnid & 0xFF)
	plmnBytes[1] = uint8((plmnid >> 8) & 0xFF)
	plmnBytes[2] = uint8((plmnid >> 16) & 0xFF)

	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, plmnBytes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeCIDDecStrToUint64 decodes the CID dec string type to bytes type
func DecodeCIDDecStrToUint64(cidDecStr string) (uint64, error) {
	cid, err := strconv.ParseUint(cidDecStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return cid, err
}

// DecodePlmnIDHexStrToBytes decodes the PLMNID hex string type to bytes type
func DecodePlmnIDHexStrToBytes(plmnidHexStr string) ([]byte, error) {
	var plmnBytes [3]uint8
	n, err := strconv.ParseUint(plmnidHexStr, 16, 32)
	if err != nil {
		return nil, err
	}
	plmnid := uint32(n)

	plmnBytes[0] = uint8(plmnid & 0xFF)
	plmnBytes[1] = uint8((plmnid >> 8) & 0xFF)
	plmnBytes[2] = uint8((plmnid >> 16) & 0xFF)

	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, plmnBytes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeCIDHexStrToUint64 decodes the CID string type to bytes type
func DecodeCIDHexStrToUint64(cidDecStr string) (uint64, error) {
	cid, err := strconv.ParseUint(cidDecStr, 16, 64)
	if err != nil {
		return 0, err
	}
	return cid, err
}


