package monitor

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Writer 应当可以从错误中恢复, 不应该对监控系统 甚至业务产生影响
// Writer 应当负责格式化数据并以对应的方式最终处理
// 如: 格式化为 普罗米修斯 的方式,上传 or 落地 及 其它 ?
// 如: 格式化为 Txt 的方式,上传 or 落地 及 其它 ?
type Writer interface {
	DoWithRecover(nameMap *MetricNameMap, omd *OneMinStorage) error // 数据写入方法
	// ISMe(name string) bool                                      // 应当具有一个指标名, 可判断是否进行http格式化
	// HTTPFormat(omd *OneMinStorage, name string) ([]byte, error) // HTTP 方式格式化数据,返回 []byte, 不执行其它处理 ???
}

const (
	// ALL 模式 all 的写法
	ALL = 0
	// AllStr 模式 all 的写法
	AllStr = "all"

	// UP 模式 up 的写法
	UP = 1
	// UpStr 模式 up 的写法
	UpStr = "up"

	// DOWN 模式 down 的写法
	DOWN = 2
	// DownStr 模式 down 的写法
	DownStr = "down"

	// ModeError 常量字符串, 描述正确的 writer 模式
	ModeError = `mode must in (int(0, 1, 2) or string("all", "up", "down") not case sensitive)`

	// UPloadRetryError 常量字符串, 描述正确的上传重试次数
	UPloadRetryError = `Upload Retry in 0-10, if n>10 please use force=true argument.`
)

var (
	// NewLineBytes 用于时间戳后的空行
	NewLineBytes = []byte("\n\n")

	// BaseSuffix 基础类型的后缀,默认求和 同 Sum
	BaseSuffix = []string{"_Sum"}

	// SumSuffix Sum类型的后缀,默认求和 同 Base
	SumSuffix = []string{"_Sum"}

	// CountSuffix Count 类型的后缀
	CountSuffix = []string{"_Count"}

	// AvgSuffix  Avg 类型的后缀
	AvgSuffix = []string{"_Avg"}

	// 多结尾的后缀,需要和取值一一对应 !!! 重点!!!

	// CountSumSuffix Count 和 Sum 结尾的后缀
	CountSumSuffix = []string{"_Count", "_Sum"}
	// CountAvgSuffix Count 和 Avg 结尾的后缀
	CountAvgSuffix = []string{"_Count", "_Avg"}

	// QuantileSuffix 分位数后缀
	QuantileSuffix = []string{"_MinP50", "_MinP90", "_MinP95", "_MinP99"}
)

// RegisterWriterName 存储已注册的 Writer Name
// 与 InitWriter 方法中的一一对应
var RegisterWriterName = map[string]func(conf *WriterConfig) Writer{}

// ErrorWriterConfig Writer 配置错误
type ErrorWriterConfig struct {
	Msg string // 信息
}

// Error  ErrorWriteConfig 的 error 接口实现
func (e *ErrorWriterConfig) Error() string {
	return e.Msg
}

// WriterConfig 一个 Writer 的配置信息
type WriterConfig struct {
	Name        string // 定义一个名字, 用于 http 格式化的时候是否需要自己进行
	Mode        int    // 工作模式
	UpLoadHost  string // 上传的主机地址 ip or 主机名
	UpLoadPort  int    // 上传的主机的端口
	UploadRetry int    // 上传失败重试次数
	DownPath    string // 落地文件的路径及名字, 需要注意冲突及权限, 建议相对路径
}

// NewWriterConfig 返回一个默认的 writer 配置
// 默认模式为落地本地
// 路径为 "./go-monitor.txt"
func NewWriterConfig() *WriterConfig {
	return &WriterConfig{
		Name:        "",
		Mode:        2,
		UpLoadHost:  "",
		UpLoadPort:  2003,
		UploadRetry: 0,
		DownPath:    "./go-monitor.txt",
	}
}

// ValidWriterName 校验 Writer 的 Name 是否合法
// 参考 RegisterWriterName 和 InitWriter
func (w *WriterConfig) ValidWriterName(name string) error {
	if _, ok := RegisterWriterName[name]; ok {
		w.Name = name
		return nil
	}
	return &ErrorWriterConfig{Msg: "Can't find This Writer"}
}

// ValidateMode 模式配置校验及配置
func (w *WriterConfig) ValidateMode(mode interface{}) error {
	// 默认模式
	mi := 9
	ok := false

	// 如果输入的是 int 则直接校验返回
	if mi, ok = mode.(int); ok {
		return w.ValidateIntMode(mi)
	}

	// 如果为 string 类型将其转换为 int 类型再次校验
	if ms, ok := mode.(string); ok {
		switch strings.ToLower(ms) {
		case AllStr:
			mi = 0
		case UpStr:
			mi = 1
		case DownStr:
			mi = 2
		}

		return w.ValidateIntMode(mi)
	}

	return &ErrorWriterConfig{Msg: ModeError}
}

// ValidateIntMode writer 工作模式校验
func (w *WriterConfig) ValidateIntMode(mode int) error {
	if mode == 0 || mode == 1 || mode == 2 {
		w.Mode = mode
		return nil
	}
	return &ErrorWriterConfig{Msg: ModeError}
}

// ValidateUpLoadHost 模式配置校验及配置
func (w *WriterConfig) ValidateUpLoadHost(host string) error {
	if w.Mode == DOWN {
		return nil
	}
	// 可能需要其它校验逻辑,暂未想到很多  不能为空???  TODO
	return nil
}

// ValidateUpLoadPort 模式配置校验及配置
func (w *WriterConfig) ValidateUpLoadPort(port int) error {
	if w.Mode == DOWN {
		return nil
	}
	// 校验
	if port < 0 || port > 65535 {
		return &ErrorWriterConfig{Msg: "Upload Port must be > 0 and < 65535"}
	}
	// 应用
	w.UpLoadPort = port
	return nil
}

// ValidateUploadRetry 上传重试次数校验
func (w *WriterConfig) ValidateUploadRetry(n int, force bool) error {
	if n < 0 {
		return &ErrorWriterConfig{Msg: UPloadRetryError}
	}

	if n > 10 && !force {
		return &ErrorWriterConfig{Msg: UPloadRetryError}
	}

	w.UploadRetry = n
	return nil
}

// ValidateDownPath 模式配置校验及配置
func (w *WriterConfig) ValidateDownPath(path string) error {
	// 可能需要其它校验逻辑,暂未想到很多  不能为空???  TODO
	return nil
}

// getValueString 格式化 特殊监控的 value 值, 并返回
// 本方法不支持格式化分位数的值, 原因在于分位数可能需要返回多个
func getValueString(_type int, SPV *SpecValue) []string {
	switch _type {
	case BaseMetric:
		return []string{strconv.FormatFloat(SPV.Sum, 'f', 5, 64)}
	case SumMetric:
		return []string{strconv.FormatFloat(SPV.Sum, 'f', 5, 64)}
	case AvgMetric:
		return []string{strconv.FormatFloat(SPV.Sum/float64(SPV.Count), 'f', 5, 64)}
	case CountMetric:
		return []string{strconv.FormatInt(SPV.Count, 10)}
	case CountSumMetric:
		return []string{
			strconv.FormatInt(SPV.Count, 10),
			strconv.FormatFloat(SPV.Sum, 'f', 5, 64),
		}
	case CountAvgMetric:
		return []string{
			strconv.FormatInt(SPV.Count, 10),
			strconv.FormatFloat(SPV.Sum/float64(SPV.Count), 'f', 5, 64),
		}
	case QuantileMetric:
		return []string{
			strconv.FormatFloat(SPV.Otd.Quantile(0.50), 'f', 5, 64),
			strconv.FormatFloat(SPV.Otd.Quantile(0.90), 'f', 5, 64),
			strconv.FormatFloat(SPV.Otd.Quantile(0.95), 'f', 5, 64),
			strconv.FormatFloat(SPV.Otd.Quantile(0.99), 'f', 5, 64),
		}
	}

	return []string{""}
}

// getTempFile 获取一个临时文件
func getTempFile() (*os.File, error) {
	return ioutil.TempFile("/tmp", "TempGoMonitorText_")
}

// InitWriter 初始化一个 writer
// 注意 Name 和 Writer 的对应关系
// 新增加其它Writer类型应该在 ValidWriterName 中进行添加
func InitWriter(conf *WriterConfig) (Writer, error) {
	if f, ok := RegisterWriterName[conf.Name]; ok {
		w := f(conf)
		return w, nil
	}
	return nil, &ErrWriterNotFound{Name: conf.Name}
}
