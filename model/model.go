package model

import (
	"encoding/json"
	"log"
	"time"
)

type Throttle struct {
	// | 0 | Under-voltage detected |
	CurrentlyUndervoltageDetected bool `json:"currentlyUndervoltageDetected"`

	// | 1 | Arm frequency capped |
	ArmFrequencyCapped bool `json:"armFrequencyCapped"`

	// | 2 | Currently throttled |
	CurrentlyThrottled bool `json:"currentlyThrottled"`

	// | 3 | Soft temperature limit active |
	CurrentlySofTempratureLimitActive bool `json:"currentlySofTempratureLimitActive"`

	// | 16 | Under-voltage has occurred |
	UndervoltageOccurred bool `json:"undervoltageOccurred"`

	// | 17 | Arm frequency capped has occurred |
	ArmFrequencyCappedOccurred bool `json:"armFrequencyCappedOccurred"`

	// | 18 | Throttling has occurred |
	ThrottlingOccured bool `json:"throttlingOccured"`

	// | 19 | Soft temperature limit has occurred
	SofTempratureLimitOccurred bool `json:"sofTempratureLimitOccurred"`
}

type Value struct {
	Temperature *float64  `json:"temperature,omitempty"`
	Throttle    *Throttle `json:"throttle,omitempty"`
}

type Entry struct {
	Datetime int64   `json:"datetime"`
	Value    *Value  `json:"value,omitempty"`
	Error    *string `json:"error,omitempty"`
}

func NewValueEntry(v *Value) *Entry {
	return &Entry{
		Datetime: time.Now().Unix(),
		Value:    v,
		Error:    nil,
	}
}

func NewErrorEntry(e error) *Entry {
	es := new(string)
	*es = e.Error()

	return &Entry{
		Datetime: time.Now().Unix(),
		Value:    nil,
		Error:    es,
	}
}

func (e *Entry) Save() {
	b, _ := json.Marshal(e)
	log.Println(string(b))
}

func NewThrottleFromInt(throttle int) *Throttle {
	// | 0 | Under-voltage detected |
	// | 1 | Arm frequency capped |
	// | 2 | Currently throttled |
	// | 3 | Soft temperature limit active |
	// | 16 | Under-voltage has occurred |
	// | 17 | Arm frequency capped has occurred |
	// | 18 | Throttling has occurred |
	// | 19 | Soft temperature limit has occurred
	return &Throttle{
		CurrentlyUndervoltageDetected:     (throttle & 1) == 1,
		ArmFrequencyCapped:                (throttle & 2) == 2,
		CurrentlyThrottled:                (throttle & 4) == 4,
		CurrentlySofTempratureLimitActive: (throttle & 8) == 8,
		UndervoltageOccurred:              (throttle & (1 << 16)) == (1 << 16),
		ArmFrequencyCappedOccurred:        (throttle & (1 << 17)) == (1 << 17),
		ThrottlingOccured:                 (throttle & (1 << 18)) == (1 << 18),
		SofTempratureLimitOccurred:        (throttle & (1 << 19)) == (1 << 19),
	}
}
