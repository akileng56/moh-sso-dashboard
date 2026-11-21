package reportscheduler_test

import (
	"fmt"
	"time"

	scheduler "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

// External consumers can use the service through its public worker boundary.
var _ scheduler.WorkerService = (*scheduler.Service)(nil)

func ExampleCalculateNextRun() {
	after := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)
	next, err := scheduler.CalculateNextRun("daily", "Africa/Kampala", scheduler.ScheduleTiming{TimeOfDay: "08:00"}, after)
	if err != nil {
		panic(err)
	}
	fmt.Println(next.Format(time.RFC3339))
	// Output: 2026-11-18T05:00:00Z
}
