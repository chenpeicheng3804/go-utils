package util

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	Files     []string   // 筛选后的文件列表
	mu        sync.Mutex // 用于保护files列表的互斥锁
	condition string     // 筛选文件的条件
}

// NewFileProcessor 创建一个新的FileProcessor实例。
// condition: 筛选文件的正则表达式条件。
// 返回值: 初始化后的FileProcessor指针。
func NewFileProcessor(condition string) *FileProcessor {
	return &FileProcessor{
		Files:     []string{},
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

			f.Files = append(f.Files, path)
		}

		return nil
	}
}

// 逐行读取文件内容 返回字符串切片
func ReadFile(filePath string) ([]string, error) {
	// // 打开文件
	// file, err := os.Open(filePath)
	// if err != nil {
	// 	return nil, err
	// }
	// defer file.Close()

	// // 创建一个字符串切片用于存储文件内容
	// var lines []string

	// reader := bufio.NewReader(file)

	// // 循环读取文件内容
	// for {
	// 	re := regexp.MustCompile(`# .*\n|-- .*\n`)
	// 	lineTmp, _, err := reader.ReadLine()
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	lineTmp = bytes.TrimPrefix(lineTmp, []byte{0xef, 0xbb, 0xbf})
	// 	line := re.ReplaceAllString(string(lineTmp), "")
	// 	line = strings.TrimSpace(line)
	// 	if line == "" {
	// 		continue
	// 	}
	// 	lines = append(lines, line)
	// }

	// return lines, nil

	var lines []string

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Println(filePath, "文件不存在:", err)
		return lines, err
	}
	// 读取SQL文件内容，并忽略错误
	readBytes, _ := os.ReadFile(filePath)
	readBytesTmp := bytes.TrimPrefix(readBytes, []byte{0xef, 0xbb, 0xbf})
	// 将SQL文件内容按分号分割成数组
	readArr := strings.Split(string(readBytesTmp)+"\n", ";\n")
	// 创建正则表达式，用于匹配SQL注释
	re := regexp.MustCompile(`# .*\n|-- .*\n`)
	for _, line := range readArr {
		// 使用正则表达式替换SQL中的注释
		line = re.ReplaceAllString(line, "")
		// 去除SQL语句两端的空白字符
		line = strings.TrimSpace(line)
		// 如果SQL为空，则跳过本次循环
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}
