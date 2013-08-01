package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type endToEndTest struct {
	madefile []string
	change   []string
	expected []string
	execs    []string
}

func testEndToEnd(t *testing.T, tcase endToEndTest) {
	tmp, stop := TempDir()
	defer stop()

	check(ioutil.WriteFile(
		tmp+madeFile,
		[]byte(strings.Join(tcase.madefile, "\n")),
		0777))

	initExecs, err := Made(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(initExecs) > 0 {
		t.Fatal("Should not execute anything on init")
	}

	createAll(tmp, tcase.change)

	excs, err := Made(tmp)
	if err != nil {
		t.Fatal(err)
	}

	for i, name := range tcase.change {
		if strings.HasSuffix(name, "/") {
			tcase.change[i] = name[:len(name)-1]
		}
	}
	built, missing := filesetDiff(listAll(tmp), tcase.change)
	for _, m := range missing {
		t.Fail()
		t.Error("Made deleted:", m)
	}
	compareFilesets(t, built, tcase.expected)

	fmt.Println(excs)
	var i int
	var ex Execution
	for i, ex = range excs {
		if len(tcase.execs) > i {
			if tcase.execs[i] != ex.Cmd {
				t.Error("Executed:", ex.Cmd, "instead of:", tcase.execs[i])
			}
		} else {
			t.Error("Unexpected executed:", ex.Cmd)
		}
	}
	i++
	if len(tcase.execs) > i {
		for _, c := range tcase.execs[i:] {
			t.Error("Did not execute:", c)
		}
	}
}

func TestSimpleCommand(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |a| b"},
		change:   []string{"a", "c"},
		expected: []string{"b"},
		execs:    []string{"cp a b"},
	})
}

func TestMultipleCommands(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |a| b", "cp |c| d"},
		change:   []string{"a", "c"},
		expected: []string{"b", "d"},
		execs:    []string{"cp a b", "cp c d"},
	})
}

func TestDependendCommands(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |a| b", "cp |b| c"},
		change:   []string{"a", "d"},
		expected: []string{"b", "c"},
		execs:    []string{"cp a b", "cp b c"},
	})
}

func TestClassicPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |a*| b"},
		change:   []string{"a", "d"},
		expected: []string{"b"},
		execs:    []string{"cp a* b"},
	})
}

func TestForEachPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |*.a| %b"},
		change:   []string{"a.a", "d"},
		expected: []string{"a.ab"},
		execs:    []string{"cp a.a a.ab"},
	})
}

func TestCopyToDirectory(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp |a| dir"},
		change:   []string{"a", "dir/"},
		expected: []string{"dir/a"},
		execs:    []string{"cp a dir"},
	})
}

func TestDoNothing(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp c d"},
		change:   []string{"a", "b"},
		expected: []string{},
		execs:    []string{},
	})
}

func compareFilesets(t *testing.T, left, right []string) {
	for _, l := range left {
		found := false
		for i, r := range right {
			if l == r {
				found = true
				right[i] = ""
			}
		}
		if !found && l != stateFile[1:] && l != madeFile[1:] {
			t.Fail()
			t.Error("Made created:", l)
		}
	}
	for _, r := range right {
		if r != "" {
			t.Fail()
			t.Error("Made did not create:", r)
		}
	}
}

func filesetDiff(left, right []string) (res, missing []string) {
	for _, r := range right {
		found := false
		for i, l := range left {
			if l == r {
				left[i] = ""
				found = true
			}
		}
		if !found {
			missing = append(missing, r)
		}
	}
	for _, l := range left {
		if l != "" {
			res = append(res, l)
		}
	}
	return
}

func createAll(base string, fps []string) {
	for _, fp := range fps {
		if strings.Contains(fp, "/") {
			if fp[len(fp)-1] == '/' {
				check(os.MkdirAll(filepath.Join(base, fp), 0777))
				continue
			} else {
				check(os.MkdirAll(filepath.Dir(filepath.Join(base, fp)), 0777))
			}
		}
		check(ioutil.WriteFile(
			filepath.Join(base, fp), []byte(""), 0777))
	}
}

func TempDir() (string, func()) {
	path, err := ioutil.TempDir("", "TempDir")
	check(err)
	path, err = filepath.EvalSymlinks(path)
	check(err)
	return path, func() {
		check(os.RemoveAll(path))
	}
}
