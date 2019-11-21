package monitor

import (
	"bytes"
	"os"
	"strconv"
)

// 落地为 Text 的 writer

// init 注册一个初始化 TextWriter 的 Writer
func init() {
	f := func(conf *WriterConfig) Writer {
		return &TextWriter{
			Conf:     conf,
			Describe: "A Text type writer",
		}
	}

	RegisterWriterName["TextWriter"] = f
}

const (
	// NewLine LF 换行
	NewLine = byte(10)
	// Equal 等号 用于分割tag的 key value, 分割值和value
	Equal = byte(61)
	// Semicolon 分号, 用于 Watcher 时候分割 Tag的相关
	Semicolon = byte(59)
)

// TextWriter 一个 Text writer
// 落地为本地文件
type TextWriter struct {
	Conf     *WriterConfig
	Describe string
}

// DoWithRecover 处理一分钟的数据
func (j *TextWriter) DoWithRecover(nameMap *MetricNameMap, omd *OneMinStorage) error {
	defer func() {
		if p := recover(); p != nil {
			Logger.Printf("TextWriter Have Panic at DoWithRecover %s", p)
		}
	}()

	// 初始化数据文件
	tmpfile, err := getTempFile()
	if err != nil {
		Logger.Printf("TextWriter Get Temp File Fail, Error %s", err)
		return err
	}
	tmpfile.WriteString(strconv.FormatInt(omd.Ts.Unix(), 10))
	tmpfile.Write(NewLineBytes)

	// 加锁
	nameMap.RLock()
	defer nameMap.RUnlock()
	omd.RLock()
	defer omd.RUnlock()

	// 特殊监控值写入到临时文件中
	err = writeTextSpecMetrics(tmpfile, nameMap.Map, omd.PersistentData)
	if err != nil {
		Logger.Printf("TextWriter Write Spec Metric on Temp File Error %s", err)
		return err
	}
	// 普通的监控数据写入到临时文件
	err = writeTextBasicMetrics(tmpfile, omd.Data)
	if err != nil {
		Logger.Printf("TextWriter Write Basic Metric on Temp File Error %s", err)
		return err
	}

	// 关闭文件 重命名
	tmpfile.Close()
	err = os.Rename("/tmp/"+tmpfile.Name(), j.Conf.DownPath)
	if err != nil {
		return err
	}

	return nil
}

// 普通的监控数据写入到临时文件
func writeTextBasicMetrics(f *os.File, data map[string]float64) error {
	var buf bytes.Buffer

	for name, value := range data {
		buf.Reset()
		buf.WriteString(name)
		buf.WriteByte(Equal)
		buf.WriteString(strconv.FormatFloat(value, 'f', 5, 65))
		buf.WriteByte(NewLine)
		_, err := f.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

// 特殊监控值写入到临时文件中
func writeTextSpecMetrics(f *os.File, nameMap map[int]*MetricName, data map[int]*SpecValue) error {
	var buf bytes.Buffer

	for id, SPV := range data {
		buf.Reset()
		_name := nameMap[id].Name
		_tags := nameMap[id].GetSortedTags()

		switch nameMap[id].Type {
		case BaseMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(BaseMetric, SPV))
		case SumMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(SumMetric, SPV))
		case AvgMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(AvgMetric, SPV))
		case CountMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(CountMetric, SPV))
		case CountSumMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(CountSumMetric, SPV))
		case CountAvgMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(CountAvgMetric, SPV))
		case QuantileMetric:
			writeTextstringSlice(f, buf, &_name, _tags, CountSuffix, getValueString(QuantileMetric, SPV))
		}
	}
	return nil
}

// writeTextstringSlice 将具有多个后缀和多个value的数据写入文件
func writeTextstringSlice(f *os.File, buf bytes.Buffer, baseName *string, _tags [][]byte, suffixes []string, value []string) (count int, err error) {
	for i, v := range value {
		newName := *baseName + suffixes[i]
		writeTextMetric(buf, &newName, _tags, v)
	}
	return f.Write(buf.Bytes())
}

// writeMetric 写入只有一个值的数据到value
func writeTextMetric(buf bytes.Buffer, _name *string, _tags [][]byte, value string) {
	buf.WriteString(*_name)
	if len(_tags) > 0 {
		for _, kv := range _tags {
			buf.WriteByte(Semicolon)
			buf.Write(kv)
		}
	}
	buf.WriteByte(Equal)
	buf.WriteString(value)
	buf.WriteByte(NewLine)
}
