// siztree loop through directories and determine file sizes.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"
)

type node struct {
	parent   *node
	info     os.FileInfo
	path     string
	children []*node
}

var root string
var wg sync.WaitGroup

var rootNode = new(node)
var (
	level = flag.Int("level", 1, "depth of size to print tree")
	help  = flag.Bool("h", false, "print usage")
)

var writer *tabwriter.Writer

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	writer = new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 0, '\t', 0)

	var err error
	root, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	wg.Add(1)
	walk(root, rootNode)
	wg.Wait()
	var rootSize int64
	format(rootNode, &rootSize, *level)

	writer.Flush()
}

func walk(path string, parent *node) {
	defer wg.Done()

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if subpath == root {
			rootNode.path = subpath
			rootNode.info = info
		}

		next := &node{
			info:   info,
			parent: parent,
			path:   subpath,
		}

		if info.IsDir() && path != subpath {
			wg.Add(1)

			parent.children = append(parent.children, next)
			go walk(subpath, next)
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() {
			parent.children = append(parent.children, next)
		}

		return nil
	})

	if err != nil {
		log.Println(err)
	}
}

func format(n *node, size *int64, level int) {
	var nextS int64
	mySize := n.info.Size()
	for _, sn := range n.children {
		format(sn, &nextS, level-1)
	}
	mySize += nextS
	*size += mySize
	if level >= 0 {
		fmt.Fprintf(writer, "%s\t%s\n", n.path, byteCountIEC(mySize))
	}
}

// src: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// src: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
