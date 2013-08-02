package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var _ = fmt.Fscan

const (
	stateFile = "/.made.json"
	madeFile  = "/Madefile"
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
		// fmt.Println(fk, tv, fv)
		if ok {
			delete(cTo, fk)
			if tv.After(fv) {
				compared[fk] = true
			}
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
		writeState(root, to)
		return
	}
	check(err)
	check(json.Unmarshal(b, &from))
	writeState(root, to)
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

// %f - whole filepath
// %d - dirname of path
// %b - basename of file
// %B - basename without extension
// %e - extension of file
func replaceForEach(name, arg string) (string, bool) {
	li := -1
	found := false
	for i := 0; i > li; {
		li = i
		if i = strings.Index(arg, "%"); i > -1 {
			ext := filepath.Ext(name)
			if len(ext) > 1 {
				ext = ext[1:] // strip
			}
			base := filepath.Base(name)
			m := map[rune]string{
				'f': name,
				'd': filepath.Dir(name),
				'b': base,
				'B': strings.TrimSuffix(base, filepath.Ext(name)),
				'e': ext,
			}
			i++
			r := rune(arg[i])
			insert, ok := m[r]
			if ok {
				found = true
				arg = strings.Replace(arg, "%"+string(r), insert, -1)
				// println(arg, insert, strings.TrimSuffix("dir/a.a", filepath.Ext("dir/a.a")))
			}
		} else {
			break
		}
	}
	return arg, found
}

func filterCmds(change string, rules [][]string) (cmds [][]string) {
	// fmt.Println("change: ", change)
	for _, rule := range rules {
		cmd := append([]string{}, rule...)
		// fmt.Println(cmd, rule)
		// fmt.Println("Processing Cmd:", cmd, "with change: ", change)
		foundPercent := false
		matchedIndices := []int{}
		for i, arg := range rule {
			// println(arg, change)
			if strings.HasPrefix(arg, "|") && strings.HasSuffix(arg, "|") {
				arg = arg[1 : len(arg)-1]
				cmd[i] = arg
				m, _ := filepath.Match(arg, change)
				if m {
					matchedIndices = append(matchedIndices, i)
				}
			}
			cmd[i], foundPercent = replaceForEach(change, arg)
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
	return cmds
}

func Made(root string) (excs []Execution, err error) {
	for i := 0; i < 4; i++ {
		_, changed := getChanged(root)
		var mf []byte
		mf, err = ioutil.ReadFile(root + madeFile)
		if err != nil {
			return
		}
		// if strings.Contains(string(mf), "\n") {
		// 	fmt.Println(changed)
		// }

		didExec := false
		for change := range changed {
			cmds := parseMadefile(mf)
			cmds = filterCmds(change, cmds)

			for _, cmd := range cmds {
				excs = append(excs, run(root, cmd))
				didExec = true
			}
		}
		if !didExec {
			return
		}
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
