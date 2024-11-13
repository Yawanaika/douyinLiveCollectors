package time

import (
	"douyinLiveCollectors/backend/common/enums"
	"time"
)

func ParseEventTime(u uint64) string {
	return time.Unix(int64(u), 0).Format(enums.TimeFormat)
}

func Now() string {
	return time.Now().Format(enums.TimeFormat)
}

func Today() string {
	return time.Now().Format(enums.TimeDayFormat)
}
