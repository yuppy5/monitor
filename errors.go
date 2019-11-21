package monitor

import "fmt"

// 各种 Error 的 集合

// ErrorLoggerLevel 定义一个错误的日志
type ErrorLoggerLevel struct{}

func (e *ErrorLoggerLevel) Error() string {
	return "Level Must in [DEBUG, INFO, WARN, ERROR]"
}

// ErrMonitorConfig 错误的配置
type ErrMonitorConfig struct {
	Msg string
}

// Error 实现 error 接口
func (e ErrMonitorConfig) Error() string {
	return fmt.Sprintf("Config Error: %s", e.Msg)
}

// ErrWriteText 写入 Text 异常
type ErrWriteText struct {
	Msg interface{}
}

// Error error 接口实现
func (e *ErrWriteText) Error() string {
	return fmt.Sprintf("%v", e.Msg)
}

// ErrWriterNotFound 找不到的 Writer
// 可能未注册
type ErrWriterNotFound struct {
	Name string // Writer 名, 注册时候使用, 应当为唯一标识, 否则可能覆盖
}

// Error 实现 error 接口
func (e ErrWriterNotFound) Error() string {
	return fmt.Sprintf("Writer name :%s not found, may be not register", e.Name)
}

// ErrNotFoundMetric 指标 id 不存在的错误
type ErrNotFoundMetric struct {
	ID   int    // 映射的指标id
	Name string // 指标名
}

func (n *ErrNotFoundMetric) Error() string {
	return fmt.Sprintf("Not found Metric MapID: %d, Name: %s", n.ID, n.Name)
}

// ErrUnexpectMetricType 定义之外的指标类型
type ErrUnexpectMetricType struct{}

func (u *ErrUnexpectMetricType) Error() string {
	return "Unexpect Metric Type."
}

// ErrHTTPPort http 端口错误
type ErrHTTPPort struct{}

func (e *ErrHTTPPort) Error() string {
	return "Http Port Must be >1024 and  < 65535"
}
