package goimport

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func ParseRelation(
	rootPath, seekPath, excludeFile string, leafVisibility, includeTests bool) *ImportPathFactory {

	factory := NewImportPathFactory(
		rootPath,
		seekPath,
		excludeFile,
		leafVisibility,
		includeTests,
	)
	factory.Root = factory.Get(rootPath)
	if factory.Root == nil {
		return nil
	}
	return factory

}

type ImportPathFactory struct {
	Root         *ImportPath
	Filter       *ImportFilter
	Pool         map[string]*ImportPath
	excludeFile  string
	includeTests bool
}

func NewImportPathFactory(
	rootPath, seekPath, excludeFile string, leafVisibility, includeTests bool) *ImportPathFactory {

	if rootPath == "." {

	}
	self := &ImportPathFactory{
		Pool:         make(map[string]*ImportPath),
		excludeFile:  excludeFile,
		includeTests: includeTests,
	}
	filter := NewImportFilter(
		rootPath,
		seekPath,
		leafVisibility,
	)
	self.Filter = filter
	return self
}
func (self *ImportPathFactory) GetRoot() *ImportPath {
	return self.Root
}

func (self *ImportPathFactory) GetAll() []*ImportPath {
	ret := make([]*ImportPath, 0)
	for _, value := range self.Pool {
		ret = append(ret, value)
	}
	return ret
}

func (self *ImportPathFactory) Get(importPath string) *ImportPath {
	// aquire from pool
	pool := self.Pool
	if _, ok := pool[importPath]; ok {
		return pool[importPath]
	}
	filter := self.Filter
	// if not applicable return nullobject
	if !filter.Applicable(importPath) {
		// if invisible return nil
		if !filter.Visible(importPath) {
			return nil
		}
		pool[importPath] = &ImportPath{
			ImportPath: importPath}
		return pool[importPath]
	}

	dirPath, err := goSrc(importPath)
	if err != nil {
		// if invisible return nil
		if !filter.Visible(importPath) {
			return nil
		}
		pool[importPath] = &ImportPath{
			ImportPath: importPath}
		return pool[importPath]
	}
	ret := &ImportPath{
		ImportPath: importPath,
	}
	pool[importPath] = ret
	fileNames := glob(dirPath, self.excludeFile, self.includeTests)
	ret.Init(self, fileNames)
	return ret
}

//ImportFilter
type ImportFilter struct {
	root     string
	seekPath string
	plotLeaf bool
}

func NewImportFilter(root string, seekPath string, plotLeaf bool) *ImportFilter {
	if seekPath == "SELF" {
		seekPath = root
	}
	impf := &ImportFilter{
		root:     root,
		seekPath: seekPath,
		plotLeaf: plotLeaf,
	}
	return impf

}

func (self *ImportFilter) Visible(path string) bool {
	return self.plotLeaf
}

func (self *ImportFilter) Applicable(path string) bool {
	if self.seekPath == "" {
		return true
	}
	if strings.Index(path, self.seekPath) == 0 {
		return true
	}
	return false
}

func isMatched(pattern string, target string) bool {
	r, _ := regexp.Compile(pattern)
	return r.MatchString(target)
}

func glob(dirPath, excludeFile string, includeTests bool) []string {
	fileNames, err := filepath.Glob(filepath.Join(dirPath, "/*.go"))
	if err != nil {
		panic("no gofiles")
	}

	files := make([]string, 0, len(fileNames))

	for _, v := range fileNames {
		if !includeTests && isMatched("_test[.]go", v) {
			continue
		}
		if !includeTests && isMatched("_example[.]go", v) {
			continue
		}
		p := filepath.Join(dirPath, v)
		if excludeFile != "" && isMatched(excludeFile, p) {
			continue
		}
		files = append(files, v)
	}
	return files
}

var errBuiltin = errors.New("Built-in package")
var goSrcResult = map[string]string{}
var goSrcError = map[string]error{
	"C": errBuiltin,
}

func goSrc(importPath string) (string, error) {
	if result, ok := goSrcResult[importPath]; ok {
		return result, nil
	}
	if err, ok := goSrcError[importPath]; ok {
		return "", err
	}
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", importPath)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		goSrcError[importPath] = err
		return "", err
	}
	result := strings.TrimSpace(string(out))
	goSrcResult[importPath] = result
	return result, nil
}
