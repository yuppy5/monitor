package monitor

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	err = os.Rename(tmpfile.Name(), j.Conf.DownPath)
	if err != nil {
		// Logger.Printf("TextWriter Write Rename Temp File Error %s", err)
		fmt.Printf("TextWriter Write Rename Temp File Error %s", err)
		return err
	}

	return nil
}

// 特殊监控值写入到临时文件中
func writeTextSpecMetrics(f *os.File, nameMap map[int]*MetricName, data map[int]*SpecValue) (err error) {
	var buf strings.Builder

	for id, SPV := range data {
		_name := nameMap[id].Name
		_tags := GetSortedTagsString(nameMap[id].GetSortedTags())
		values := getValueString(nameMap[id].Type, SPV)
		suffix := SuffixMap[nameMap[id].Type]

		for i, value := range values {
			fmt.Fprintf(&buf, "%s%s%s=%s\n", _name, suffix[i], _tags, value)
		}
	}

	_, err = f.WriteString(buf.String())
	return
}

// GetSortedTagsString 格式化tag 字符串并返回
func GetSortedTagsString(SortedTags [][]byte) (ret string) {
	if len(SortedTags) == 0 {
		return
	}

	var buf bytes.Buffer

	for _, tag := range SortedTags {
		buf.WriteByte(Semicolon)
		buf.Write(tag)
	}

	ret = buf.String()

	return
}

// 普通的监控数据写入到临时文件
func writeTextBasicMetrics(f *os.File, data map[string]float64) (err error) {
	var buf strings.Builder

	for name, value := range data {
		fmt.Fprintf(&buf, "%s=%f\n", name, value)
	}

	_, err = f.WriteString(buf.String())
	return
}
