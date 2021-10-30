package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileNode struct {
	Name string
	Size int64
}

func (f FileNode) IsDir() bool {
	return false
}

func (f FileNode) GetName() string {
	return f.Name
}

func (f FileNode) RenderNode(depth int, isLast bool) []string {
	var size string
	if f.Size > 0 {
		size = fmt.Sprintf("%db", f.Size)
	} else {
		size = "empty"
	}

	line := fmt.Sprintf("%s%s (%s)", getIndent(depth, isLast), f.GetName(), size)
	return []string{line}
}

type DirNode struct {
	Name     string
	Children []Node
}

func (d DirNode) IsDir() bool {
	return true
}

func (d DirNode) GetName() string {
	return d.Name
}

func (d DirNode) RenderNode(depth int, isLast bool) []string {
	lines := []string{getIndent(depth, isLast) + d.GetName()}

	subLines := renderAst(d.Children, depth+1)

	for idx, subLine := range subLines {
		if !isLast {
			subLines[idx] = "│\t" + subLine
		} else {
			subLines[idx] = "\t" + subLine
		}
	}

	return append(lines, subLines...)
}

type Node interface {
	IsDir() bool
	GetName() string
	RenderNode(depth int, isLast bool) []string
}

func buildAst(path string, parentPath string, printFiles bool) []Node {
	file, err := os.Open(filepath.Join(parentPath, path))

	if err != nil {
		panic(err)
	}

	elements, err := file.ReadDir(0)

	ast := make([]Node, 0)

	for _, el := range elements {
		if el.IsDir() {
			ast = append(ast, DirNode{
				Name:     el.Name(),
				Children: buildAst(el.Name(), filepath.Join(parentPath, path), printFiles),
			})
		} else if printFiles {
			info, _ := el.Info()
			ast = append(ast, FileNode{
				Name: el.Name(),
				Size: info.Size(),
			})
		}
	}
	return ast
}

func getIndent(depth int, isLast bool) string {
	result := ""

	// indent := strings.Repeat(" ", depth*1)
	if isLast {
		result += "└───"
	} else {
		result += "├───"
	}

	return result
}

func renderFileNode(fn FileNode, depth int, isLast bool) string {
	return getIndent(depth, isLast) + fn.GetName()
}

func renderAst(ast []Node, depth int) []string {
	sort.Slice(ast, func(i, j int) bool {
		return ast[i].GetName() < ast[j].GetName()
	})

	lines := make([]string, 0)

	for idx, n := range ast {
		isLast := idx == len(ast)-1

		lines = append(lines, n.RenderNode(depth, isLast)...)
	}

	return lines
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	ast := buildAst(path, "", printFiles)

	tree := strings.Join(renderAst(ast, 0), "\n")

	fmt.Fprintln(out, tree)

	return nil
}

func main() {
	out := os.Stdout
	fmt.Println(os.Args)
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
