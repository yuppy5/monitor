package monitor

import (
	"io/ioutil"
	"log"
	"sync"
	"time"
	// _ "net/http/pprof"
)

var (
	// Logger 日志输出, 可通过 monitor.Logger 修改
	Logger StdLogger = log.New(ioutil.Discard, "[GoMonitor] ", log.LstdFlags)
)

// StdLogger is used to log error messages.
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// MONITOR 模块本身
type MONITOR struct {
	sync.RWMutex          // 保护Client状态
	Conf         *Config  // 配置文件
	Core         *Storage // 核心存储

	writers []Writer // 数据后续处理接口

	closer, closed chan struct{} // 用于关闭后台落地文件的程序 发送数据 export 等
}

// New 返回一个监控monitor 实例
func New(conf *Config) (*MONITOR, error) {
	// 无需配置校验, 因为每一次添加已经进行过校验,且默认配置为合法的
	// 基础初始化
	m := &MONITOR{}
	m.closer = make(chan struct{}, 1)
	m.closed = make(chan struct{}, 1)

	// 配置初始化
	m.Conf = conf

	// 核心初始化
	m.Core = NewStorage(conf.Revisions)

	for _, wc := range conf.Writers {
		writer, err := InitWriter(wc)
		if err != nil {
			Logger.Printf("Init Writer Error %s", err)
			continue
		}
		m.writers = append(m.writers, writer)
	}

	return m, nil
}

// HTTPPort 校验并更新配置中的 HTTP 端口
func (m *MONITOR) HTTPPort(port int) error {
	return m.Conf.ValidateHTTPPort(port)
}

// Interval 配置聚合周期, 单位为 s
// 聚合最小单位为 秒
func (m *MONITOR) Interval(n int) error {
	dur := time.Duration(n) * time.Second
	return m.Conf.ValidateInterval(dur)
}

// Revisions 保留历史的版本数
func (m *MONITOR) Revisions(n int) error {
	return m.Conf.ValidateRevisions(n)
}

// Start 启动监控
// 有必要增加 recover 方式启动, 以使监控系统出问题的时候不影响业务
// 启动 HTTP Listen  TODO ???
// 启动周期性的执行 NextMonitor, 并将返回结果给 []writer 处理, 需要 recover
// 周期为 60s ,按照实际分钟的 0s 开始处理, 其它直接周期处理
func (m *MONITOR) Start() {
	var t *time.Ticker

	m.StartHTTPModule(m.Conf.Port)

	if int(m.Conf.Interval.Seconds()) == 60 {
		startInWholeMinute()
		t = time.NewTicker(m.Conf.Interval)
	} else {
		t = time.NewTicker(m.Conf.Interval)
	}

	// ticker 启动前执行一次
	now := m.Core.NextMonitor()
	for _, writer := range m.writers {
		go writer.DoWithRecover(m.Core.MetricMap, now)
	}

	// 周期执行
	go func() {
		defer close(m.closed)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				// 周期性执行 NextMonitor 并将结果交给 Writer 处理
				now := m.Core.NextMonitor()
				for _, writer := range m.writers {
					go writer.DoWithRecover(m.Core.MetricMap, now)
				}
			case <-m.closer:
				Logger.Println("Monitor Recv Close Single, quit...")
				return
			}
		}
	}()
}

// Stop 停止监控.  可能采用其它逻辑 TODO
func (m *MONITOR) Stop() {
	Logger.Println("Will Stop Monitor...")
	close(m.closer)

	<-m.closed
}

// startInWholeMinute 阻塞, 直到分钟为整
func startInWholeMinute() {
	now := time.Now()
	// 计算下一个整数分钟
	next := now.Add(time.Second * 60)
	next = time.Date(next.Year(), next.Month(), next.Day(),
		next.Hour(), next.Minute(), 0, 0, next.Location())
	// 整分钟后触发, 结束函数
	t := time.NewTimer(next.Sub(now))
	<-t.C
}
