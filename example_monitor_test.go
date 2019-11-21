package monitor

import (
	"time"
)

var M *MONITOR

// curl http://127.0.0.1:9999/monitor/monitor.TRecordFuncCount
// curl http://127.0.0.1:9999/monitor/monitor.TRecordFuncTimeAvg
// curl http://127.0.0.1:9999/monitor/monitor.TRecordFuncTimes
// curl http://127.0.0.1:9999/history/444/monitor.TRecordFuncCount
// curl http://127.0.0.1:9999/history/444/monitor.TRecordFuncTimeAvg
// curl http://127.0.0.1:9999/history/444/monitor.TRecordFuncTimes

func init() {
	cfgWriter1 := NewWriterConfig()
	cfgWriter1.ValidWriterName("TextWriter")
	cfgMonitor := NewConfig()
	cfgMonitor.ValidateInterval(time.Duration(1) * time.Second)
	cfgMonitor.ValidateRevisions(10)
	cfgMonitor.AddWriter(cfgWriter1)

	var err error
	M, err = New(cfgMonitor)
	if err != nil {
		Logger.Printf("Start Monitor Error : %s", err)
	}
}

func Example() {
	M.Start()
	defer M.Stop()

	for i := 0; i < 2000; i++ {
		go TRecordFuncCount()
		go TRecordFuncTimeAvg()
		TRecordFuncTimes()
	}

	// Output:
}

func TRecordFuncCount() {
	defer M.RecordFuncCount()()
	time.Sleep(time.Duration(3) * time.Millisecond)
}

func TRecordFuncTimeAvg() {
	defer M.RecordFuncTimeAvg()()
	time.Sleep(time.Duration(5) * time.Millisecond)
}

func TRecordFuncTimes() {
	defer M.RecordFuncTimes()()
	time.Sleep(time.Duration(10) * time.Millisecond)
}
