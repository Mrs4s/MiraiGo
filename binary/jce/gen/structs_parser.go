package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "embed"
)

//go:embed structs_parser_tmp.go
var tmpmain string

func main() {
	f := flag.String("f", "structs_tobytes.go", "output file.")
	i := flag.String("i", "structs.go", "input file.")
	flag.Parse()
	fmt.Println("gen runs on arg", *f, *i)
	fmt.Println("len of tmp main is", len(tmpmain))
	tmp, err := os.Create("tmp.go")
	if err != nil {
		panic(err)
	}
	inp, err := os.Open(*i)
	if err != nil {
		panic(err)
	}
	var structs []string
	var isinhead = true
	var isinfill = false
	var isnexttemplate = false
	s := bufio.NewScanner(bytes.NewReader([]byte(tmpmain)))
	for s.Scan() {
		line := s.Text()
		if isinhead {
			if strings.Contains(line, "Main_") {
				line = "func main() {"
				isnexttemplate = true
				isinhead = false
				scanner := bufio.NewScanner(inp)
				start := false
				for scanner.Scan() {
					t := scanner.Text()
					if t == "type (" {
						start = true
					}
					if start {
						tmp.WriteString(t + "\n")
						if t == ")" {
							break
						}
						if strings.Contains(t, " struct {") {
							structs = append(structs, strings.Trim(t[:len(t)-9], "\t"))
						}
					}
				}
				inp.Close()
			}
			tmp.WriteString(line + "\n")
		} else if isinfill {
			for _, s := range structs {
				fmt.Fprintf(tmp, "\tWriteJceStruct(w, &%s{})\n", s)
			}
			isinfill = false
			tmp.WriteString("// structs_parser: fill area\n")
			tmp.WriteString(line + "\n")
		} else {
			if strings.Contains(line, "// structs_parser: fill area") {
				isinfill = true
				tmp.WriteString("// structs_parser: fill area\n")
			} else {
				if isnexttemplate {
					fmt.Fprintf(tmp, line+"\n", *f)
					isnexttemplate = false
				} else {
					tmp.WriteString(line + "\n")
				}
			}
		}
	}
	tmp.Close()
	c := exec.Command("go", "run", "tmp.go")
	err = c.Run()
	if err != nil {
		panic(err)
	}
	os.Remove("tmp.go")
}
