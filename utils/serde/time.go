package serde

import (
	"encoding/json"
	"strconv"
	"time"
)

type Time time.Time

func (j Time) MarshalJSON() ([]byte, error) {
	ms := time.Time(j).UnixMilli()
	i := strconv.FormatInt(ms, 10)
	return json.Marshal(i)
}

func (j *Time) UnmarshalJSON(data []byte) error {
	var ms int64
	err := json.Unmarshal(data, &ms)
	if err != nil {
		return err
	}

	*j = Time(time.Unix(0, ms*int64(time.Millisecond)))
	return nil
}
