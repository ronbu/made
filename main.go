package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

}

func Made(root string) (excs []Execution, err error) {
	cont, err := ioutil.ReadFile(root + "/Madefile")
	if err != nil {
		return
	}

	fps := listAll(root)
	cmds := append(parseMadefile(fps, cont))

	for _, cmd := range cmds {
		run(root, cmd)
	}

	return
}

func listAll(dir string) (files []string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		check(err)
		if path != dir {
			relPath, err := filepath.Rel(dir, path)
			check(err)
			files = append(files, relPath)
		}
		return nil
	})
	return files
}

func parseMadefile(changed []string, madeFile []byte) (cmds [][]string) {
	// Split function seems to have a bug
	fixSplit := func(s, sep string) (splitted []string) {
		splitted = strings.Split(s, sep)
		if len(splitted) == 1 && splitted[0] == "" {
			splitted = []string{}
		}
		return
	}
	// println("New mdfile: ", madeFile)
	for _, change := range changed {
		// fmt.Println("change: ", change)
		for _, line := range fixSplit(string(madeFile), "\n") {
			// line = strings.Trim(line, " \t")
			// if strings.HasPrefix(line, "#") {
			// 	continue
			// }
			foundPercent := false
			matchedIndices := []int{}
			cmd := fixSplit(line, " ")
			for i, arg := range cmd {
				// println(arg, change)
				m, _ := filepath.Match(arg, change)
				if m {
					// println("matched: ", arg)
					matchedIndices = append(matchedIndices, i)
				}
				if strings.Contains(arg, "%") {
					foundPercent = true
					cmd[i] = strings.Replace(arg, "%", filepath.Base(change), -1)
				}
			}
			if len(matchedIndices) > 0 {
				if foundPercent {
					for _, i := range matchedIndices {
						cmd[i] = change
					}
				}
				fmt.Printf("matched: %v, in cmd: %v\n", matchedIndices, cmd)
				cmds = append(cmds, cmd)
			}
		}
	}
	return
}

func run(base string, cmd []string) {
	cmdStr := cmd[0]
	for _, arg := range cmd[1:] {
		cmdStr += " " + arg
	}
	// cmdStr += "\""

	sh := "/usr/local/bin/fish"
	c := exec.Command(sh, "-c", cmdStr)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	c.Dir = base
	fmt.Printf("Executing: %#v\n", c.Args)
	err := c.Start()
	check(err)
	err = c.Wait()
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Execution struct {
	Cmd   []string
	Errs  []error
	Fatal bool
}

func (e *Execution) String() string {
	str := "Executed: "
	if e.Fatal {
		str = "Failed to Execute: "
	}
	str += strings.Join(e.Cmd, " ") + "\n"
	for _, err := range e.Errs {
		str += "\t" + err.Error() + "\n"
	}
	return str
}
