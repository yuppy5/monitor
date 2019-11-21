package monitor

import (
	"runtime"
	"time"
)

// 部分特殊指标和值初始化需要的方法

// getCallNameMaeID 获取一个映射的 ID, 存在的话 返回 id 和 true
func (m *MONITOR) getCallNameMaeID(CallName string) (id int, ok bool) {
	// 如果是第一次访问则初始化整个数据, 包括当前监控数据
	m.Core.MetricMap.RLock()

	if id, ok = m.Core.MetricMap.CallNameMap[CallName]; ok {
		m.Core.MetricMap.RUnlock()
		return
	}

	m.Core.MetricMap.RUnlock()
	return -1, false
}

// initNewMetricName 初始化一个指标的映射,并初始化当前监控中特殊类型的值
// 返回初始化是用到的 ID
func (m *MONITOR) initNewMetricName(_name string, _type int, _desc string,
	tags map[string]string) (_id int, err error) {

	// 初始化指标名 struct
	mStruct := initMetricName(_name, _type, _desc, tags)
	vStruct, err := initSpecValue(_type)

	m.Core.MetricMap.Lock()
	// 获得一个ID
	m.Core.MetricMap.LastID++
	_id = m.Core.MetricMap.LastID
	// 初始化 name 和 id 的 映射
	m.Core.MetricMap.CallNameMap[_name] = _id
	// 初始化指标名类型
	m.Core.MetricMap.Map[_id] = mStruct

	// 初始化当前监控数据
	m.Core.NowMonitor.Lock()
	m.Core.NowMonitor.PersistentData[_id] = vStruct
	m.Core.NowMonitor.Unlock()

	// map 的 锁释放放到最后, 防止当前监控还没添加上其它协程获取到 ID
	m.Core.MetricMap.Unlock()

	return
}

// Add Set AddPersistent SetPersistent  RecordFuncCount RecordFuncTimes RecordFuncTimeAvg

// Add 调用一分钟存储的 Add 实现
func (m *MONITOR) Add(name string, value float64) {
	m.Core.NowMonitor.Add(name, value)
}

// Set 调用一分钟存储的 Set 实现
func (m *MONITOR) Set(name string, value float64) {
	m.Core.NowMonitor.Set(name, value)
}

// AddPersistent 调用一分钟存储的 AddPersistent 实现
func (m *MONITOR) AddPersistent(MapID int, metricType int, value float64) {
	m.Core.NowMonitor.AddPersistent(MapID, metricType, value)
}

// SetPersistent 调用一分钟存储的 SetPersistent 实现
func (m *MONITOR) SetPersistent(MapID int, metricType int, value float64) {
	m.Core.NowMonitor.SetPersistent(MapID, metricType, value)
}

// 特殊 RecordFunc 方法

// RecordFuncTimes 记录函数的调用次数
func (m *MONITOR) RecordFuncTimes() func() {
	// 使用 runtime 等返回函数调用堆栈 然后记录, caller 参数 1 可能有问题,需测试 TODO
	start := time.Now()
	pc, _, _, _ := runtime.Caller(1)
	callFuncName := runtime.FuncForPC(pc).Name()
	return func() {
		m.Add(callFuncName, time.Now().Sub(start).Seconds()*1000)
	}
}

// RecordFuncTimeAvg 记录函数调用的平均时间消耗
// 特殊数据中记录总数, 最后落地前计算平均值
// 未使用 RecordMetircTimeAvg 作为底层是为了将获取函数名的时间也计入开销
func (m *MONITOR) RecordFuncTimeAvg() func() {
	start := time.Now()
	pc, _, _, _ := runtime.Caller(1)
	callFuncName := runtime.FuncForPC(pc).Name()

	id := -1
	id, ok := m.getCallNameMaeID(callFuncName)
	if !ok {
		tags := make(map[string]string)
		id, _ = m.initNewMetricName(callFuncName, AvgMetric,
			"Record a func time avg metric", tags)
	}

	return func() {
		m.AddPersistent(id, AvgMetric, time.Now().Sub(start).Seconds()*1000)
	}
}

// RecordMetircTimeAvg 对给定对指标求平均值
// 需要给出指标名, 除此之外,其余的都与 RecordFuncTimeAvg 相同
func (m *MONITOR) RecordMetircTimeAvg(CallName string) func() {
	start := time.Now()
	id := -1
	id, ok := m.getCallNameMaeID(CallName)
	if !ok {
		tags := make(map[string]string)
		id, _ = m.initNewMetricName(CallName, AvgMetric,
			"Record a func time avg metric", tags)
	}

	return func() {
		m.AddPersistent(id, AvgMetric, time.Now().Sub(start).Seconds()*1000)
	}
}

// RecordFuncCount 记录函数的调用次数
func (m *MONITOR) RecordFuncCount() func() {
	// 使用 runtime 等返回函数调用堆栈 然后记录, caller 参数 1 可能有问题,需测试 TODO
	pc, _, _, _ := runtime.Caller(1)
	callFuncName := runtime.FuncForPC(pc).Name()
	return func() {
		m.Add(callFuncName, 1.0)
	}
}

// RecordMetricCount 记录 metric 的次数
func (m *MONITOR) RecordMetricCount(name string) func() {
	return func() {
		m.Add(name, 1.0)
	}
}

// RecordMetricSum 记录 metric 的 和
// 底层采用 Add 方法, 仅封装用于 record 方法调用方式的统一
func (m *MONITOR) RecordMetricSum(name string, value float64) func() {
	return func() {
		m.Add(name, value)
	}
}

// RecordMetric 记录最后一次 metric 的值
// 底层采用 Set 方法, 仅封装用于 record 方法调用方式的统一
func (m *MONITOR) RecordMetric(name string, value float64) func() {
	return func() {
		m.Set(name, value)
	}
}
