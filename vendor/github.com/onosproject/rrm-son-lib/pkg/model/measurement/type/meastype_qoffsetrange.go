// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package meastype

import "fmt"

const (
	// QOffsetMinus24dB is the Q-Offset value -24 dB
	QOffsetMinus24dB QOffsetRange = iota
	// QOffsetMinus22dB is the Q-Offset value -22 dB
	QOffsetMinus22dB
	// QOffsetMinus20dB is the Q-Offset value -20 dB
	QOffsetMinus20dB
	// QOffsetMinus18dB is the Q-Offset value -18 dB
	QOffsetMinus18dB
	// QOffsetMinus16dB is the Q-Offset value -16 dB
	QOffsetMinus16dB
	// QOffsetMinus14dB is the Q-Offset value -14 dB
	QOffsetMinus14dB
	// QOffsetMinus12dB is the Q-Offset value -12 dB
	QOffsetMinus12dB
	// QOffsetMinus10dB is the Q-Offset value -10 dB
	QOffsetMinus10dB
	// QOffsetMinus8dB is the Q-Offset value -8 dB
	QOffsetMinus8dB
	// QOffsetMinus6dB is the Q-Offset value -6 dB
	QOffsetMinus6dB
	// QOffsetMinus5dB is the Q-Offset value -5 dB
	QOffsetMinus5dB
	// QOffsetMinus4dB is the Q-Offset value -4 dB
	QOffsetMinus4dB
	// QOffsetMinus3dB is the Q-Offset value -3 dB
	QOffsetMinus3dB
	// QOffsetMinus2dB is the Q-Offset value -2 dB
	QOffsetMinus2dB
	// QOffsetMinus1dB is the Q-Offset value -1 dB
	QOffsetMinus1dB
	// QOffset0dB is the Q-Offset value 0 dB
	QOffset0dB
	// QOffset1dB is the Q-Offset value 1 dB
	QOffset1dB
	// QOffset2dB is the Q-Offset value 2 dB
	QOffset2dB
	// QOffset3dB is the Q-Offset value 3 dB
	QOffset3dB
	// QOffset4dB is the Q-Offset value 4 dB
	QOffset4dB
	// QOffset5dB is the Q-Offset value 5 dB
	QOffset5dB
	// QOffset6dB is the Q-Offset value 6 dB
	QOffset6dB
	// QOffset8dB is the Q-Offset value 8 dB
	QOffset8dB
	// QOffset10dB is the Q-Offset value 10 dB
	QOffset10dB
	// QOffset12dB is the Q-Offset value 12 dB
	QOffset12dB
	// QOffset14dB is the Q-Offset value 14 dB
	QOffset14dB
	// QOffset16dB is the Q-Offset value 16 dB
	QOffset16dB
	// QOffset18dB is the Q-Offset value 18 dB
	QOffset18dB
	// QOffset20dB is the Q-Offset value 20 dB
	QOffset20dB
	// QOffset22dB is the Q-Offset value 22 dB
	QOffset22dB
	// QOffset24dB is the Q-Offset value 24 dB
	QOffset24dB
)

// NewQOffsetRange returns NewQOffsetRange object
func NewQOffsetRange(qOffset QOffsetRange) MeasType {
	if qOffset < QOffsetMinus24dB || qOffset > QOffset24dB {
		return &DefaultQOffsetRange
	}
	return &qOffset
}

// QOffsetRange is the type for Q-Offset range
type QOffsetRange int

// SetValue is the set function for value
func (q *QOffsetRange) SetValue(i interface{}) error {
	if i.(QOffsetRange) < QOffsetMinus24dB || i.(QOffsetRange) > QOffset24dB {
		*q = DefaultQOffsetRange
		return fmt.Errorf("Q-Offset should be set in the range QOFFSET_MINUS24DB to QOFFSET_24DB; received %v - set to default QOFFSET_0DB", i.(QOffsetRange))
	}
	*q = i.(QOffsetRange)
	return nil
}

// GetValue is the get function for value as interface type
func (q *QOffsetRange) GetValue() interface{} {
	return valueListQOffsetRange[*q]
}

// String returns value as string type
func (q *QOffsetRange) String() string {
	return strListQOffsetRange[*q]
}

// DefaultQOffsetRange is the default value of Q-Offset range
var DefaultQOffsetRange = QOffset0dB

var strListQOffsetRange = []string{
	"QOffsetMinus24dB",
	"QOffsetMinus22dB",
	"QOffsetMinus20dB",
	"QOffsetMinus18dB",
	"QOffsetMinus16dB",
	"QOffsetMinus14dB",
	"QOffsetMinus12dB",
	"QOffsetMinus10dB",
	"QOffsetMinus8dB",
	"QOffsetMinus6dB",
	"QOffsetMinus5dB",
	"QOffsetMinus4dB",
	"QOffsetMinus3dB",
	"QOffsetMinus2dB",
	"QOffsetMinus1dB",
	"QOffset0dB",
	"QOffset1dB",
	"QOffset2dB",
	"QOffset3dB",
	"QOffset4dB",
	"QOffset5dB",
	"QOffset6dB",
	"QOffset8dB",
	"QOffset10dB",
	"QOffset12dB",
	"QOffset14dB",
	"QOffset16dB",
	"QOffset18dB",
	"QOffset20dB",
	"QOffset22dB",
	"QOffset24dB",
}

var valueListQOffsetRange = []int{
	-24, //"QOffsetMinus24dB"
	-22, //"QOffsetMinus22dB"
	-20, //"QOffsetMinus20dB"
	-18, //"QOffsetMinus18dB"
	-16, //"QOffsetMinus16dB"
	-14, //"QOffsetMinus14dB"
	-12, //"QOffsetMinus12dB"
	-10, //"QOffsetMinus10dB"
	-8,  //"QOffsetMinus8dB"
	-6,  //"QOffsetMinus6dB"
	-5,  //"QOffsetMinus5dB"
	-4,  //"QOffsetMinus4dB"
	-3,  //"QOffsetMinus3dB"
	-2,  //"QOffsetMinus2dB"
	-1,  //"QOffsetMinus1dB"
	0,   //"QOffset0dB"
	1,   //"QOffset1dB"
	2,   //"QOffset2dB"
	3,   //"QOffset3dB"
	4,   //"QOffset4dB"
	5,   //"QOffset5dB"
	6,   //"QOffset6dB"
	8,   //"QOffset8dB"
	10,  //"QOffset10dB"
	12,  //"QOffset12dB"
	14,  //"QOffset14dB"
	16,  //"QOffset16dB"
	18,  //"QOffset18dB"
	20,  //"QOffset20dB"
	22,  //"QOffset22dB"
	24,  //"QOffset24dB"
}
