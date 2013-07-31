package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	stateFile = "/.made.json"
)

type files map[string]time.Time
type changes map[string]bool

func main() {

}

func mapFiles(root string) (fs files) {
	fs = make(files)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		check(err)
		if path != root {
			relPath, err := filepath.Rel(root, path)
			check(err)
			fs[relPath] = info.ModTime()
		}
		return nil
	})
	return
}

func writeState(root string, files files) {
	b, err := json.Marshal(files)
	check(err)
	ioutil.WriteFile(root+stateFile, b, 0777)
}

func compareFileMaps(from, to files) (compared changes) {
	cTo := make(files)
	compared = make(changes)
	for k, v := range to {
		cTo[k] = v
	}
	for fk, fv := range from {
		tv, ok := to[fk]
		if ok && tv.After(fv) {
			compared[fk] = true
			delete(cTo, fk)
		} else {
			compared[fk] = false
		}
	}
	for k, _ := range cTo {
		compared[k] = true
	}
	return
}

func getChanged(root string) (init bool, changed changes) {
	from := make(files)
	changed = make(changes)
	to := mapFiles(root)
	b, err := ioutil.ReadFile(root + stateFile)
	if os.IsNotExist(err) {
		for k := range to {
			changed[k] = true
		}
		return true, changed
	}
	check(err)
	check(json.Unmarshal(b, &from))
	return init, compareFileMaps(from, to)
}

func run(root string, cmd []string) Execution {
	cmdStr := cmd[0]
	for _, arg := range cmd[1:] {
		cmdStr += " " + arg
	}
	sh := "/usr/local/bin/fish"
	c := exec.Command(sh, "-c", cmdStr)
	// c.Stderr = os.Stderr
	// c.Stdout = os.Stdout
	c.Dir = root
	output, err := c.CombinedOutput()
	return Execution{
		Cmd:    cmdStr,
		Err:    err,
		Output: output}
}

func filterCmds(changed changes, rules [][]string) (cmds [][]string) {
	for change := range changed {
		// fmt.Println("change: ", change)
		for _, rule := range rules {
			cmd := append([]string{}, rule...)
			// fmt.Println(cmd, rule)
			// fmt.Println("Processing Cmd:", cmd, "with change: ", change)
			foundPercent := false
			matchedIndices := []int{}
			for i, arg := range rule {
				// println(arg, change)
				m, _ := filepath.Match(arg, change)
				if m {
					// println("matched: ", arg)
					matchedIndices = append(matchedIndices, i)
				}
				if strings.Contains(arg, "%") {
					foundPercent = true
					cmd[i] = strings.Replace(arg, "%", filepath.Base(change), -1)
					// fmt.Printf("Found Percent in: %v, at arg: %s\n", rule, arg)
				}
			}
			if len(matchedIndices) > 0 {
				if foundPercent {
					for _, i := range matchedIndices {
						cmd[i] = change
					}
				}
				// fmt.Printf("matched: %v, in cmd: %v, with change: %s\n", matchedIndices, rule, change)
				cmds = append(cmds, cmd)
			}
		}
	}
	return cmds
}

func Made(root string) (excs []Execution, err error) {
	// for {
	var exs []Execution
	_, changed := getChanged(root)
	// if len(changed) == 0 {
	// 	return
	// }
	var madefile []byte
	madefile, err = ioutil.ReadFile(root + "/Madefile")
	if err != nil {
		return
	}
	cmds := parseMadefile(madefile)
	cmds = filterCmds(changed, cmds)

	for _, cmd := range cmds {
		exs = append(exs, run(root, cmd))
	}

	excs = append(excs, exs...)
	if len(exs) == 0 {
		return
	}
	// }
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

func parseMadefile(madeFile []byte) (cmds [][]string) {
	// Split function seems to have a bug
	fixSplit := func(s, sep string) (splitted []string) {
		splitted = strings.Split(s, sep)
		if len(splitted) == 1 && splitted[0] == "" {
			splitted = []string{}
		}
		return
	}
	// println("New mdfile: ", madeFile)
	for _, line := range fixSplit(string(madeFile), "\n") {
		// line = strings.Trim(line, " \t")
		// if strings.HasPrefix(line, "#") {
		// 	continue
		// }
		cmds = append(cmds, fixSplit(line, " "))
	}
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Execution struct {
	Cmd    string
	Err    error
	Output []byte
}

func (e *Execution) String() string {
	output := string(e.Output)
	str := "Executed: "
	if e.Err != nil {
		str = "Failed to Execute: "
	}
	str += e.Cmd
	if e.Err != nil {
		str += "\n\t" + e.Err.Error()
	}
	if output != "" {
		str += "\n\t" + string(e.Output)
	}
	return str
}
