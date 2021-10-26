package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var (
	s, t string
)

func main() {
	flag.StringVar(&s, "s", "", "move source")
	flag.StringVar(&t, "t", "", "move target")
	flag.Parse()

	if isNotExist(s) {
		log.Printf("Source not found;\n")
		return
	}
	if isNotExist(t) {
		log.Printf("Target not found;\n")
		return
	}
	if isDir, _ := isDir(t); !isDir {
		log.Printf("%s must be dir;\n", t)
		return
	}

	var wg sync.WaitGroup
	ch := make(chan struct{}, 100)
	var count int32

	start := time.Now()
	filepath.Walk(s, func(path string, info os.FileInfo, err error) error {
		isDir, err := isDir(path)
		if err != nil {
			return err
		}
		if isDir {
			return nil
		}
		go func() {
			wg.Add(1)
			ch <- struct{}{}
			err = move(path, info)
			log.Printf("%s complete.", path)
			atomic.AddInt32(&count, 1)
			if err == nil {
				log.Printf("Deleting %s ", path)
				os.Remove(path)
			}
			<-ch
			wg.Done()
		}()
		return nil
	})
	wg.Wait()
	fmt.Printf("Move %d files, total used %.2f minites.", count, time.Since(start).Minutes())
}

func isNotExist(path string) bool {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		log.Println(path)
		return true
	}
	return false
}

func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		return true, nil
	}
	return false, nil
}

func move(s string, fi os.FileInfo) error {
	fileName := fi.Name()

	// create file
	newFile, err := os.Create(filepath.Join(t, fileName))
	defer newFile.Close()
	if err != nil {
		log.Printf("Create %s error.\n", filepath.Join(t, fileName))
		return err
	}

	originFile, err := os.Open(s)

	defer originFile.Close()
	if err != nil {
		log.Printf("Open %s error.\n", s)
		return err
	}
	_, err = io.Copy(newFile, originFile)
	if err != nil {
		log.Printf("Copy %s error %v.\n", s, err)
		return err
	}
	return nil
}
