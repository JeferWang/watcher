package watcher

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//定义错误
var (
	ErrDurationTooShort = errors.New("间隔时间不能小于1毫秒")
	ErrWatcherRunning   = errors.New("Watcher已经在运行")
)

//定义结构:操作
type Op uint32

//初始化需要使用的Op常亮值
const (
	//创建文件
	Create Op = iota
	//写&修改文件
	Write
	//删除
	Remove
	//重命名
	Rename
	//修改权限
	Chmod
	//移动
	Move
)

//定义结构:Event
type Event struct {
	Op
	Path string
	os.FileInfo
}

//todo 定义结构：Watcher

//todo 定义函数:Watcher实例化

//todo 定义函数:Watcher添加文件

//todo 定义函数：读取目录所有文件

type Watcher struct {
	//=====事件=====
	Event  chan Event    //对外发送文件变更事件事件
	Error  chan error    // 对外发送
	Closed chan struct{} //对外发送停止Watcher事件
	close  chan struct{} //接收内部close事件
	//=====数据=====
	watched  map[string]bool        //需要监控的文件、目录列表
	fileList map[string]os.FileInfo //监控的文件列表
	running  bool                   // 运行状态标记位
	//=====同步=====
	wg *sync.WaitGroup //保证多任务是线性执行的
	mu *sync.Mutex     //锁
}

// 实例化一个Watcher
func New() *Watcher {
	var wg sync.WaitGroup
	wg.Add(1)
	return &Watcher{
		Event:    make(chan Event),
		Error:    make(chan error),
		Closed:   make(chan struct{}),
		close:    make(chan struct{}),
		watched:  make(map[string]bool),
		fileList: make(map[string]os.FileInfo),
		running:  false,
		wg:       &wg,
		mu:       new(sync.Mutex),
	}
}

//启动Watcher
func (w *Watcher) Start(d time.Duration) error {
	// 判断间隔时间是否太短
	if d < time.Nanosecond {
		return ErrDurationTooShort
	}
	// 防止重复启动
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return ErrWatcherRunning
	}
	w.running = true
	w.mu.Unlock()
	// 解除占用
	w.wg.Done()

	for {
		done := make(chan struct{})
		evt := make(chan Event)

	}
}

// 获取一个目录下所有文件和文件信息
func (w *Watcher) list(readPath string) (map[string]os.FileInfo, error) {
	fileList := make(map[string]os.FileInfo)
	err := filepath.Walk(readPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//if !info.IsDir() {
		fileList[path] = info
		//}
		return nil
	})
	return fileList, err
}

// 给Watcher添加一个目录下的所有文件
func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	fileList, err := w.list(path)
	if err != nil {
		return err
	}
	for k, v := range fileList {
		w.fileList[k] = v
	}
	w.watched[path] = false
	return nil
}

func (w *Watcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	delete(w.watched, path)
	info, found := w.fileList[path]
	if !found {
		return nil
	}
	if !info.IsDir() {
		delete(w.fileList, path)
		return nil
	}
	delete(w.fileList, path)
	for p := range w.fileList {
		if strings.HasPrefix(p, path) {
			delete(w.fileList, p)
		}
	}
	return nil
}
