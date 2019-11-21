package monitor

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// 完整已有指标访问方式, 应该需要格式化, 可采用 writer中的方式格式化后输出 or 其它

var (
	// HostName 主机名
	HostName = getHostName()
	// WEL 欢迎语
	WEL = fmt.Sprintf("WelCome Go Monitor\nHostName: %s\n", HostName)
)

// getHostName 获取主机名
func getHostName() string {
	host, err := os.Hostname()
	if err != nil {
		return "Unknown!"
	}
	return host
}

// StartHTTPModule 开始 http 模块
func (m *MONITOR) StartHTTPModule(port int) error {
	if port < 1024 || port > 65535 {
		return &ErrHTTPPort{}
	}

	ListenPortStr := ":" + strconv.Itoa(port)

	r := mux.NewRouter()
	r.HandleFunc("/", Welcome) //设置访问的路由

	r.HandleFunc("/monitor/{metric}", m.HandleMonitor).Methods("GET") //设置访问的路由

	r.HandleFunc("/history/{HVersion}/{metric}", m.HandleHistory).Methods("GET") //设置访问的路由

	go func() {
		err := http.ListenAndServe(ListenPortStr, r) //设置监听的IP和端口
		if err != nil {
			Logger.Printf("Start Http Module Error %s", err)
		}
	}()

	Logger.Println("Start Monitor Http Module Done.")

	return nil
}

// Welcome 欢迎页
func Welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, WEL)
}

// HandleMonitor http handle 用于直接http 展示 monitor
func (m *MONITOR) HandleMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Connection", "close")

	// 获取 url 中的变量
	k := vars["metric"]

	// 数据加锁
	m.Core.NowMonitor.RLock()
	defer m.Core.NowMonitor.RUnlock()

	// 输出该数据的时间
	fmt.Fprintf(w, WEL)
	fmt.Fprintf(w, fmt.Sprintf("%s\n\n", m.Core.NowMonitor.Ts.Local().String()))

	// 输出普通数据中的key
	if v, ok := m.Core.NowMonitor.Data[k]; ok {
		fmt.Fprintf(w, fmt.Sprintf("%s : %f", k, v))
	}

	// 加锁
	m.Core.MetricMap.RLock()
	defer m.Core.MetricMap.RUnlock()

	// 返回特殊类型中的 key
	if intK, ok := m.Core.MetricMap.CallNameMap[k]; ok {
		name := m.Core.MetricMap.Map[intK]
		if v, ok := m.Core.NowMonitor.PersistentData[intK]; ok {
			fmt.Fprintf(w, fmt.Sprintf("Key: %s\n%s%s", k, name.String(), v.String()))
		}
	}
}

// HandleHistory http handle 用于直接http 展示 历史 monitor
func (m *MONITOR) HandleHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Connection", "close")

	// 获取 url 中的 变量
	k := vars["metric"]
	hv, err := strconv.Atoi(vars["HVersion"])
	if err != nil {
		fmt.Fprintf(w, "History Version Type Error")
	}

	// 获取历史版本的一分钟数据
	m.Core.RLock()
	cur := m.Core.Cursor
	idx := cur - int(math.Abs(float64(hv)))
	for idx < 0 {
		idx = idx + m.Core.HistoryVersionNumber
	}
	hd := m.Core.HistoryMonitor[idx]
	m.Core.RUnlock()
	if hd == nil {
		return
	}

	// 加锁
	hd.RLock()
	defer hd.RUnlock()

	// 输出该数据的时间
	fmt.Fprintf(w, WEL)
	fmt.Fprintf(w, fmt.Sprintf("%s\n\n", hd.Ts.Local().String()))

	// 返回普通数据中的key
	if v, ok := hd.Data[k]; ok {
		fmt.Fprintf(w, fmt.Sprintf("%s : %f", k, v))
	}

	// 加锁
	m.Core.MetricMap.RLock()
	defer m.Core.MetricMap.RUnlock()

	// 返回特殊类型中的 key
	if intK, ok := m.Core.MetricMap.CallNameMap[k]; ok {
		name := m.Core.MetricMap.Map[intK]
		if v, ok := hd.PersistentData[intK]; ok {
			fmt.Fprintf(w, fmt.Sprintf("Key: %s\n%s%s", k, name, v))
		}
	}
}
