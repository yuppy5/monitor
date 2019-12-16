package monitor

import "time"

// Config Monitor 的配置文件
// 定义其工作模式, 历史版本, 聚合周期等等
type Config struct {
	// 聚合的周期, 默认 60s 一分钟
	Interval time.Duration

	// 历史监控结果保留的版本数
	Revisions int

	// 基础的监控的端口
	Port int

	// writers 配置
	Writers []*WriterConfig

	// WebPath 采用 web 方式访问的时候, 读取的文件的 path
	// 默认为 text_writer 写出来的文本 ./go-monitor.txt
	WebPath string
}

// NewConfig 返回一个 Config实例,及一些默认的配置
func NewConfig() *Config {
	return &Config{
		Interval:  60 * time.Second,
		Revisions: 3,
		Port:      9999,
		Writers:   make([]*WriterConfig, 0),
		WebPath:   "./go-monitor.txt",
	}
}

// ValidateRevisions 校验历史版本留存数量
// 保留的历史数据版本数, 默认值为 3
func (c *Config) ValidateRevisions(num int) error {
	if num < 0 {
		c.Revisions = 0
		return ErrMonitorConfig{Msg: "Revisions Conf error, use default 0"}
	}
	c.Revisions = num
	return nil
}

// ValidateInterval 校验聚合的周期
// 默认聚合周期为 60s, 建议聚合的单位为 s
func (c *Config) ValidateInterval(duration time.Duration) error {
	if duration.Seconds() < 1.0 {
		return &ErrMonitorConfig{Msg: "monitor aggregate Interval to small, must >= 1s"}
	}
	c.Interval = duration
	return nil
}

// ValidateHTTPPort 校验监听端口
// 默认监听的端口为 9999
func (c *Config) ValidateHTTPPort(port int) error {
	if c.Port == port {
		return nil
	}

	if port > 1024 && port < 65535 {
		c.Port = port
		return nil
	}

	return &ErrMonitorConfig{
		Msg: "Http Port must > 1024 and < 65535",
	}
}

// AddWriter 添加 writer 配置
func (c *Config) AddWriter(wc *WriterConfig) {
	c.Writers = append(c.Writers, wc)
}
