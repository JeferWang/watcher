package watcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

var (
	ErrSkip = errors.New("error: skipping file")
)

type Op uint32

//定义各种操作对应的数字
const (
	Create Op = iota
	Write
	Remove
	Rename
	Chmod
	Move
)

//定义一个操作常量与名称的映射
var ops = map[Op]string{
	Create: "CREATE",
	Write:  "WRITE",
	Remove: "REMOVE",
	Rename: "RENAME",
	Chmod:  "CHMOD",
	Move:   "MOVE",
}

//Op结构的输出方式
func (e Op) String() string {
	if op, isFound := ops[e]; isFound {
		return op
	}
	return "UnknownOperate"
}

//定义一个事件结构
type Event struct {
	Op   Op
	Path string
	os.FileInfo
}

//事件的字符串输出方式
func (e Event) String() string {
	if e.FileInfo == nil {
		return "UnKnownFile"
	}
	pathType := "FILE"
	if e.IsDir() {
		pathType = "DIRECTORY"
	}
	return fmt.Sprintf("%s %q %s [%s]", pathType, e.Name(), e.Op, e.Path)
}

//过滤文件的钩子函数结构
type FilterFileHookFunc func(info os.FileInfo, fullPath string) error

//正则表达式文件过滤器
func RegexFilterHook(r *regexp.Regexp, useFullPath bool) FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()
		if useFullPath {
			str = fullPath
		}
		if r.MatchString(str) {
			return nil
		}
		return ErrSkip
	}
}

type Watcher struct {
	Event  chan Event
	Error  chan error
	Closed chan struct{}
	close  chan struct{}
	wg     *sync.WaitGroup
	//mu protects the following
	mu           *sync.Mutex
	ffh          []FilterFileHookFunc
	running      bool
	names        map[string]bool        //是否递归
	files        map[string]os.FileInfo //文件映射
	ignored      map[string]struct{}    //忽略的文件或目录
	ops          map[Op]struct{}        //Op过滤
	ignoreHidden bool                   //是否忽略隐藏文件
	maxEvents    int                    //每个周期最大事件数
}

func New() *Watcher {
	var wg sync.WaitGroup
	wg.Add(1)
	return &Watcher{
		Event:   make(chan Event),
		Error:   make(chan error),
		Closed:  make(chan struct{}),
		close:   make(chan struct{}),
		wg:      &wg,
		mu:      new(sync.Mutex),
		files:   make(map[string]os.FileInfo),
		ignored: make(map[string]struct{}),
		names:   make(map[string]bool),
	}
}

func (w *Watcher) SetMaxEvent(delta int) {
	w.mu.Lock()
	w.maxEvents = delta
	w.mu.Unlock()
}
func (w *Watcher) AddFilterHook(f FilterFileHookFunc) {
	w.mu.Lock()
	w.ffh = append(w.ffh, f)
	w.mu.Unlock()
}

func (w *Watcher) IgnoreHiddenFile(ignore bool) {
	w.mu.Lock()
	w.ignoreHidden = ignore
	w.mu.Unlock()
}

func (w *Watcher) FilterOps(ops ...Op) {
	w.mu.Lock()
	w.ops = make(map[Op]struct{})
	for _, op := range ops {
		w.ops[op] = struct{}{}
	}
	w.mu.Unlock()
}
func (w *Watcher) Add(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	name, err := filepath.Abs(name)
	if err != nil {
		return err
	}
	_, ignored := w.ignored[name]
	isHidden, err := IsHiddenFile(name)
	if err != nil {
		return err
	}
	if ignored || (w.ignoreHidden && isHidden) {
		return nil
	}
	fileList, err := w.list(name)
	if err != nil {
		return err
	}
	for k, v := range fileList {
		w.files[k] = v
	}
	w.names[name] = false
	return nil
}
func (w *Watcher) list(name string) (map[string]os.FileInfo, error) {
	fileList := make(map[string]os.FileInfo)
	stat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	fileList[name] = stat
	//如果传入的路径只是一个文件，直接返回
	if !stat.IsDir() {
		return fileList, nil
	}
	fInfoList, err := ioutil.ReadDir(name)
	if err != nil {
		return nil, err
	}

	// 遍历目录下的内容
outer:
	for _, fInfo := range fInfoList {
		path := filepath.Join(name, fInfo.Name())
		_, ignored := w.ignored[path]
		isHidden, err := IsHiddenFile(path)
		if err != nil {
			return nil, err
		}
		if ignored || (isHidden && w.ignoreHidden) {
			continue //忽略当前文件（已在忽略列表，或忽略隐藏文件条件下）
		}
		//执行每一个过滤函数
		for _, f := range w.ffh {
			err := f(fInfo, path)
			//过滤器函数返回了跳过？
			if err == ErrSkip {
				continue outer //跳出当前循环
			}
		}
	}
}
