// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package ocnstorage

import (
	"github.com/onosproject/cco-mon/pkg/store/store/storage"
	"github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
)

// Entry is a struct to have IDs and Inner entry
type Entry struct {
	Key   storage.IDs
	Value InnerEntry
}

// InnerEntry is an entry of inner store element
type InnerEntry struct {
	Key   storage.IDs
	Value meastype.QOffsetRange
}

// OcnMap is the struct to store Ocn values
type OcnMap struct {
	Value map[storage.IDs]meastype.QOffsetRange
}
