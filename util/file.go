package util

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// 定义了一系列与文件处理相关的错误类型，便于更细致的错误处理。
var (
	ErrEmptyPath      = errors.New("path cannot be empty")      // 路径不能为空
	ErrEmptyCondition = errors.New("condition cannot be empty") // 条件不能为空
	ErrPathNotExist   = errors.New("path does not exist")       // 路径不存在
)

// 假设的File类型定义，增加了Mutex保护和更明确的字段名称
type FileProcessor struct {
	files     []string   // 筛选后的文件列表
	mu        sync.Mutex // 用于保护files列表的互斥锁
	condition string     // 筛选文件的条件
}

// NewFileProcessor 创建一个新的FileProcessor实例。
// condition: 筛选文件的正则表达式条件。
// 返回值: 初始化后的FileProcessor指针。
func NewFileProcessor(condition string) *FileProcessor {
	return &FileProcessor{
		files:     []string{},
		condition: condition,
	}
}

// ErgodicPathFile 遍历指定目录下的所有文件，并根据条件筛选出文件。
// Path: 需要遍历的目录路径。
// Condition: 筛选文件的正则表达式条件。
// 返回值: 筛选后的FileProcessor实例和可能发生的错误。
func ErgodicPathFile(Path, Condition string) (f *FileProcessor, Err error) {
	// 验证输入
	if Path == "" {
		return nil, ErrEmptyPath
	}
	if Condition == "" {
		return nil, ErrEmptyCondition
	}

	// 检查路径是否存在
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		return nil, ErrPathNotExist
	}

	f = NewFileProcessor(Condition)
	// 对Condition进行正则表达式预编译，优化性能
	conditionRegex, err := regexp.Compile(Condition)
	if err != nil {
		return nil, err // 如果Condition不是有效的正则表达式，返回错误
	}
	// 执行文件路径遍历
	err = filepath.Walk(Path, f.WalkFuncWithRegex(conditionRegex))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// WalkFuncWithRegex 是FileProcessor的成员方法，提供一个根据正则表达式筛选文件的filepath.WalkFunc实现。
// conditionRegex: 用于匹配文件路径的正则表达式。
// 返回值: 实现了filepath.WalkFunc接口的函数，用于遍历文件系统并筛选文件。
func (f *FileProcessor) WalkFuncWithRegex(conditionRegex *regexp.Regexp) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking the path %s: %v", path, err)
			return err
		}
		// 如果是文件且符合筛选条件，则加入到文件列表中
		if !info.IsDir() && conditionRegex.MatchString(path) {
			f.mu.Lock() // 加锁以保证并发安全
			defer f.mu.Unlock()

			f.files = append(f.files, path)
		}

		return nil
	}
}
