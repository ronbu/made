package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func run(root string, cmd string) Execution {
	sh := "/usr/local/bin/fish"
	c := exec.Command(sh, "-c", cmd)
	// c.Stderr = os.Stderr
	// c.Stdout = os.Stdout
	c.Dir = root
	output, err := c.CombinedOutput()
	return Execution{
		Cmd:    cmd,
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

func filterCmds(change, madeFile string) (cmds []string) {
	regex := regexp.MustCompile(`<\|(.+?)\|>`)
	for _, rule := range strings.Split(madeFile, "\n") {
		matched := false
		ms := regex.FindAllStringSubmatchIndex(rule, -1)
		for _, sm := range ms {
			glob := rule[sm[2]:sm[3]]
			if matched, _ = filepath.Match(glob, change); matched {
				break
			}
		}
		if !matched {
			continue
		}

		cmd, isForeach := replaceForEach(change, rule)
		if isForeach {
			cmd = regex.ReplaceAllString(cmd, change)
		} else {
			cmd = regex.ReplaceAllString(cmd, "$1")
		}
		cmds = append(cmds, cmd)
	}
	return
}

func Made(root string) (chan Execution, chan bool, error) {
	excs := make(chan Execution)
	stop := make(chan bool)
	change := make(chan string)

	init := make(chan bool)
	go func() {
		var stopped bool
		go func() {
			<-stop
			stopped = true
		}()
		lastRound := false
		for !lastRound {
			if stopped {
				lastRound = true
			}
			_, initChanges := getChanged(root)
			<-init
			for c := range initChanges {
				change <- c
			}
		}
		close(change)
	}()
	init <- true
	close(init)

	go func() {
		for change := range change {
			// LOOP:
			// 	for i := 0; i < 3; i++ {
			mf, err := ioutil.ReadFile(root + madeFile)
			if err != nil {
				return
			}
			cmds := filterCmds(change, string(mf))
			for _, c := range cmds {
				excs <- run(root, c)
			}
			// if len(cmds) == 0 {
			// 	break LOOP
			// }
			// }
		}
		close(excs)
	}()
	return excs, stop, nil
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
		str += "\n\t" + output
	}
	return str
}
