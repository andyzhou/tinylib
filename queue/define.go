package queue

import "time"

//inter macro define
const (
	DefaultQueueSize          = 1024
	DefaultListConsumeRate    = 1 //xx seconds
	DefaultGcRate             = 30
	DefaultTenThousandPercent = 10000
	DefaultTickTimer          = time.Second
)
