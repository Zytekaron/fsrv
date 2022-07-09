package serde

import (
	"encoding/json"
	"strconv"
	"time"
)

type Duration time.Duration

func (j Duration) MarshalJSON() ([]byte, error) {
	ms := time.Duration(j).Milliseconds()
	i := strconv.FormatInt(ms, 10)
	return json.Marshal(i)
}

func (j *Duration) UnmarshalJSON(data []byte) error {
	var ms int64
	err := json.Unmarshal(data, &ms)
	if err != nil {
		return err
	}

	*j = Duration(ms * int64(time.Millisecond))
	return nil
}
