// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package meastype

import "fmt"

const (
	// TTT0ms is TimeToTrigger value 0ms
	TTT0ms TimeToTriggerRange = iota
	// TTT40ms is TimeToTrigger value 0ms
	TTT40ms
	// TTT64ms is TimeToTrigger value 0ms
	TTT64ms
	// TTT80ms is TimeToTrigger value 0ms
	TTT80ms
	// TTT100ms is TimeToTrigger value 0ms
	TTT100ms
	// TTT128ms is TimeToTrigger value 0ms
	TTT128ms
	// TTT160ms is TimeToTrigger value 0ms
	TTT160ms
	// TTT256ms is TimeToTrigger value 0ms
	TTT256ms
	// TTT320ms is TimeToTrigger value 0ms
	TTT320ms
	// TTT480ms is TimeToTrigger value 0ms
	TTT480ms
	// TTT512ms is TimeToTrigger value 0ms
	TTT512ms
	// TTT640ms is TimeToTrigger value 0ms
	TTT640ms
	// TTT1024ms is TimeToTrigger value 0ms
	TTT1024ms
	// TTT1280ms is TimeToTrigger value 0ms
	TTT1280ms
	// TTT2560ms is TimeToTrigger value 0ms
	TTT2560ms
	// TTT5120ms is TimeToTrigger value 0ms
	TTT5120ms
)

// NewTimeToTriggerRange returns TimeToTriggerRange object
func NewTimeToTriggerRange(ttt TimeToTriggerRange) MeasType {
	if ttt < TTT0ms || ttt > TTT5120ms {
		return &DefaultTimeToTriggerRange
	}
	return &ttt
}

// TimeToTriggerRange is the type for TimeToTrigger
type TimeToTriggerRange int

// SetValue is the set function for value
func (t *TimeToTriggerRange) SetValue(i interface{}) error {
	if i.(TimeToTriggerRange) < TTT0ms || i.(TimeToTriggerRange) > TTT5120ms {
		*t = DefaultTimeToTriggerRange
		return fmt.Errorf("TimeToTrigger should be set in the range TTT_0MS to TTT_5120MS; received %v - set to default TTT_0MS", i.(TimeToTriggerRange))
	}
	*t = i.(TimeToTriggerRange)
	return nil
}

// GetValue is the get function for value as interface type
func (t *TimeToTriggerRange) GetValue() interface{} {
	return valueListTimeToTriggerRange[*t]
}

// String returns value as string type
func (t *TimeToTriggerRange) String() string {
	return strListTimeToTriggerRange[*t]
}

// DefaultTimeToTriggerRange is the default value of TimeToTrigger
var DefaultTimeToTriggerRange = TTT0ms

var strListTimeToTriggerRange = []string{
	"TTT0ms",
	"TTT40ms",
	"TTT64ms",
	"TTT80ms",
	"TTT100ms",
	"TTT128ms",
	"TTT160ms",
	"TTT256ms",
	"TTT320ms",
	"TTT480ms",
	"TTT512ms",
	"TTT640ms",
	"TTT1024ms",
	"TTT1280ms",
	"TTT2560ms",
	"TTT5120ms",
}

var valueListTimeToTriggerRange = []int{
	0,    //TTT0ms
	40,   //TTT40ms
	64,   //TTT64ms
	80,   //TTT80ms
	100,  //TTT100ms
	128,  //TTT128ms
	160,  //TTT160ms
	256,  //TTT256ms
	320,  //TTT320ms
	480,  //TTT480ms
	512,  //TTT512ms
	640,  //TTT640ms
	1024, //TTT1024ms
	1280, //TTT1280ms
	2560, //TTT2560ms
	5120, //TTT5120ms
}
