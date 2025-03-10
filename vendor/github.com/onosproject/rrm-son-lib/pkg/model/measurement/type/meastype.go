// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package meastype

// MeasType is the interface for measurement type
type MeasType interface {
	SetValue(interface{}) error
	GetValue() interface{}
	String() string
}
