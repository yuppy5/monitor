package monitor

import (
	"fmt"
	"sync"

	td "github.com/caio/go-tdigest"
)

// 定义几种特殊类型, 可能需要对指标特殊处理
const (
	// // 最后格式化输出时 返回一个值
	// BaseMetric 基本指标
	BaseMetric = iota
	// SumMetric 累加指标 同 BaseMetric
	SumMetric
	// AvgMetric 需要计算平均值的指标
	AvgMetric
	// CountMetric 计数指标
	CountMetric

	// 最后格式化输出时 返回两个值
	// 记录 Count 和 Sum 的特殊类型
	CountSumMetric
	// 记录 Count 和 Avg 的特殊类型
	CountAvgMetric

	// 格式化输出时,可能返回多个值
	// QuantileMetric 分位数指标
	QuantileMetric
)

// 特殊指标名类型下定义及初始化等

// MetricName 单个指标的信息
type MetricName struct {
	Name       string            // 指标的名字
	Tags       map[string]string // tag的名称和 tag 值的映射
	Describe   string            // 描述信息
	Type       int               // 指标类型
	SortedTags [][]byte          // 排过序的 byte 数组, 应该每次Tags 有变化的时候重新生成
}

func (m *MetricName) String() string {
	return fmt.Sprintf("SpecName:\n\tDescribe: %s\n\tName: %s\n\tType: %d\n\tTags: %v\n",
		m.Describe, m.Name, m.Type, m.SortedTags)
}

// initMetricName 初始化指标名
func initMetricName(name string, t int, describe string, tags map[string]string) *MetricName {
	return &MetricName{
		Name:     name,
		Tags:     tags,
		Describe: describe,
		Type:     t,
	}
}

// GetSortedTags 格式化 tags , 排序并返回
func (m *MetricName) GetSortedTags() [][]byte {
	return m.SortedTags
}

// MetricNameMap 特殊指标名前的映射数据
type MetricNameMap struct {
	sync.RWMutex
	Map         map[int]*MetricName // ID 和 特殊指标的映射
	CallNameMap map[string]int      // 函数调用类 函数名和 ID 的映射
	LastID      int                 // 最后一个 ID
}

// NewMetricNameMap 返回一个新的 MetricNameMap 映射
func NewMetricNameMap() *MetricNameMap {
	return &MetricNameMap{
		Map:         make(map[int]*MetricName),
		CallNameMap: make(map[string]int),
		LastID:      0,
	}
}

// 特殊数据类型定义及初始化等,配合特殊指标名使用

// SpecValue 特殊值组合, 应该根据指标类型特别处理
type SpecValue struct {
	sync.RWMutex
	Sum   float64     // 总和
	Count int64       // 计数
	Otd   *td.TDigest // 分位数
}

func (s *SpecValue) String() string {
	if s.Otd != nil {
		return fmt.Sprintf("SpecValue:\n\tSum: %f\n\tCount: %d\n\tTdigest: %d",
			s.Sum, s.Count, s.Otd.Count())
	}

	return fmt.Sprintf("SpecValue:\n\tSum: %f\n\tCount: %d\n",
		s.Sum, s.Count)
}

// newBaseSpecValue 返回无分位数的特殊数据集
func newBaseSpecValue() *SpecValue {
	return &SpecValue{
		Sum:   0,
		Count: 0,
		Otd:   nil,
	}
}

// newQuantileSpecValue 返回初始化过的含有分位数的特殊数据类型
func newQuantileSpecValue() (*SpecValue, error) {
	ntd, err := td.New()
	if err != nil {
		Logger.Printf("Init New Tdigest Struct Error %s", err)
		return nil, err
	}

	return &SpecValue{
		Sum:   0,
		Count: 0,
		Otd:   ntd,
	}, nil
}

// initSpecValue 初始化特殊的值类型
func initSpecValue(_type int) (*SpecValue, error) {
	switch _type {
	case BaseMetric:
		return newBaseSpecValue(), nil
	case SumMetric:
		return newBaseSpecValue(), nil
	case AvgMetric:
		return newBaseSpecValue(), nil
	case CountMetric:
		return newBaseSpecValue(), nil
	case CountSumMetric:
		return newBaseSpecValue(), nil
	case CountAvgMetric:
		return newBaseSpecValue(), nil
	case QuantileMetric:
		return newQuantileSpecValue()
	}
	return nil, &ErrUnexpectMetricType{}
}
