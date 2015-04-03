package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Context struct {
	Roots []string

	// cache components
	FileList     []map[string]string
	AllFiles     map[string]string
	MissingFiles []map[string]string

	Lock sync.Mutex

	sequence uint64
}

func find_missing(currentIndex int, catalogs []map[string]string) map[string]string {
	missing := make(map[string]string)
	for i := 0; i < len(catalogs); i++ {
		if i == currentIndex {
			continue
		}
		for entry, _ := range catalogs[i] {
			_, exists := catalogs[currentIndex][entry]
			if !exists {
				missing[entry] = entry
			}
		}
	}
	return missing
}

func enumerate_files(path string) (map[string]string, error) {
	results := make(map[string]string)

	rootLen := len(path)

	filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath := curPath
		if strings.HasPrefix(curPath, path) {
			relPath = curPath[rootLen:]
		}
		results[relPath] = relPath
		return nil
	})
	return results, nil
}

func newContext() *Context {

	return &Context{
		Roots:        loadRoots(),
		FileList:     make([]map[string]string, 0),
		AllFiles:     make(map[string]string, 0),
		MissingFiles: make([]map[string]string, 0),
	}
}

func (c *Context) Number() uint64 {
	return atomic.AddUint64(&(c.sequence), uint64(1))
}

func (c *Context) clearCache() {
	c.FileList = c.FileList[0:0]
	c.AllFiles = make(map[string]string, 0)
	c.MissingFiles = c.MissingFiles[0:0]
}

func (c *Context) saveRoots() {
	var config struct {
		Roots []string
	}
	config.Roots = c.Roots
	dat, err := json.MarshalIndent(&config, "", "    ")
	if err == nil {
		f, err := os.Create("config.json")
		if err == nil {
			defer f.Close()
			f.Write(dat)
		} else {
			log.Println("Unable to write config", err)
		}
	}
}

func loadRoots() []string {
	var config struct {
		Roots []string
	}
	config.Roots = make([]string, 0)
	dat, err := ioutil.ReadFile("config.json")
	if err != nil {
		return make([]string, 0)
	}
	err = json.Unmarshal(dat, &config)
	if err != nil {
		return make([]string, 0)
	}
	log.Println("Loaded ", len(config.Roots), "entries from the config file")
	return config.Roots
}

func (c *Context) ClearCache() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.clearCache()
}

func (c *Context) AddRoot(root string) {
	fInfo, err := os.Stat(root)
	if err != nil {
		return
	}
	if !fInfo.IsDir() {
		return
	}

	tmpFileList, err := enumerate_files(root)
	if err != nil {
		return
	}
	fileList := c.FileList
	fileList = append(fileList, tmpFileList)

	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Roots = append(c.Roots, root)
	c.clearCache()
	c.FileList = fileList

	c.saveRoots()
}

func (c *Context) DelRoot(root string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	newRoots := make([]string, 0, len(c.Roots))
	for _, entry := range c.Roots {
		if entry != root {
			newRoots = append(newRoots, entry)
		}
	}
	c.Roots = newRoots
	c.clearCache()
}

func (c *Context) InContext(root, path string) bool {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for i, entry := range c.Roots {
		if entry == root {
			_, exists := c.FileList[i][path]
			return exists
		}
	}
	return false
}

// This should only be called from the output handler it does not lock
func (c *Context) UpdateCacheNL() {
	computeAllFiles := false
	computeMissing := false
	if len(c.FileList) != len(c.Roots) {
		log.Println("Updating the file lists")
		tmpFileList := make([]map[string]string, 0)
		for _, root := range c.Roots {
			tmpList, err := enumerate_files(root)
			// FIXME: how to handle errors?
			_ = err
			tmpFileList = append(tmpFileList, tmpList)
		}
		c.FileList = tmpFileList
		computeAllFiles = true
		computeMissing = true
	}

	if computeAllFiles || len(c.AllFiles) == 0 {
		log.Println("Updating all files list")
		tmpAllFiles := make(map[string]string)
		for _, tmpFileList := range c.FileList {
			for _, entry := range tmpFileList {
				tmpAllFiles[entry] = entry
			}
		}
		c.AllFiles = tmpAllFiles
	}
	if computeMissing || len(c.Roots) != len(c.MissingFiles) {
		log.Println("Updating missing files list")
		tmpMissingFiles := make([]map[string]string, 0, len(c.Roots))
		for i := 0; i < len(c.Roots); i++ {
			tmpMissingFiles = append(tmpMissingFiles, find_missing(i, c.FileList))
		}
		c.MissingFiles = tmpMissingFiles
	}
}

func (c *Context) Odd(num uint64) bool {
	if (num & uint64(1)) == 1 {
		return true
	}
	return false
}
