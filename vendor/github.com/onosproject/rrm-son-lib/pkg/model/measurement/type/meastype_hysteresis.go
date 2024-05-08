// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package meastype

import "fmt"

// NewHysteresisRange returns NewHysteresisRange object
func NewHysteresisRange(hysteresis HysteresisRange) MeasType {
	if hysteresis < 0 || hysteresis > 30 {
		return &DefaultHysteresisRange
	}
	return &hysteresis
}

// HysteresisRange is the type for hysteresis 0 dB to 30 dB integer
type HysteresisRange int

// DefaultHysteresisRange is the default value of Hysteresis
var DefaultHysteresisRange = HysteresisRange(0)

// SetValue is the set function for value
func (h *HysteresisRange) SetValue(i interface{}) error {
	if i.(HysteresisRange) < 0 || i.(HysteresisRange) > 30 {
		*h = DefaultHysteresisRange
		return fmt.Errorf("Hysteresis should be in the range from 0 to 30; received %v - set to default 0 dB", i.(HysteresisRange))
	}
	*h = i.(HysteresisRange)
	return nil
}

// GetValue is the get function for value as interface type
func (h *HysteresisRange) GetValue() interface{} {
	return *h
}

// String returns value as string type
func (h *HysteresisRange) String() string {
	return fmt.Sprintf("%d", *h)
}
