// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package meastype

import "fmt"

// NewA3OffsetRange returns A3OffsetRange object
func NewA3OffsetRange(a3Offset A3OffsetRange) MeasType {
	if a3Offset < -30 || a3Offset > 30 {
		return &DefaultA3OffsetRange
	}
	return &a3Offset
}

// A3OffsetRange is the type for A3 offset - -30 dB to 30 dB integer
type A3OffsetRange int

// DefaultA3OffsetRange is the default value of A3 offset
var DefaultA3OffsetRange = A3OffsetRange(0)

// SetValue is the set function for value
func (a *A3OffsetRange) SetValue(i interface{}) error {
	if i.(A3OffsetRange) < -30 || i.(A3OffsetRange) > 30 {
		*a = DefaultA3OffsetRange
		return fmt.Errorf("A3-Offset should be in the range from -30 to 30; received %v - set to default 0 dB", i.(A3OffsetRange))
	}
	*a = i.(A3OffsetRange)
	return nil
}

// GetValue is the get function for value as interface type
func (a *A3OffsetRange) GetValue() interface{} {
	return *a
}

// String returns value as string type
func (a *A3OffsetRange) String() string {
	return fmt.Sprintf("%d", *a)
}
