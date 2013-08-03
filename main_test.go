package main

import (
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
	changeSet := listAll(tmp)

	excs, err := Made(tmp)
	if err != nil {
		t.Fatal(err)
	}

	for i, name := range tcase.change {
		if strings.HasSuffix(name, "/") {
			tcase.change[i] = name[:len(name)-1]
		}
	}
	// e := run(tmp, "find .")
	// t.Log(e.String())
	built, missing := filesetDiff(listAll(tmp), changeSet)
	for _, m := range missing {
		t.Fail()
		t.Error("Made deleted:", m)
	}
	t.Log(built)
	compareFilesets(t, built, tcase.expected)

	// t.Log(excs)
	var i int
	var ex Execution
	for i, ex = range excs {
		if ex.Err != nil {
			t.Log(ex.String())
		}
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
		madefile: []string{"cp <|a|> b"},
		change:   []string{"a", "c"},
		expected: []string{"b"},
		execs:    []string{"cp a b"},
	})
}

func TestMultipleCommands(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a|> b", "cp <|c|> d"},
		change:   []string{"a", "c"},
		expected: []string{"b", "d"},
		execs:    []string{"cp a b", "cp c d"},
	})
}

func TestDependendCommands(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a|> b", "cp <|b|> c"},
		change:   []string{"a", "d"},
		expected: []string{"b", "c"},
		execs:    []string{"cp a b", "cp b c"},
	})
}

func TestClassicPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a*|> b"},
		change:   []string{"a", "d"},
		expected: []string{"b"},
		execs:    []string{"cp a* b"},
	})
}

func TestForEachPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/*.a|> %fa"},
		change:   []string{"dir/a.a"},
		expected: []string{"dir/a.aa"},
		execs:    []string{"cp dir/a.a dir/a.aa"},
	})
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/*.a|> %b"},
		change:   []string{"dir/a.a"},
		expected: []string{"a.a"},
		execs:    []string{"cp dir/a.a a.a"},
	})
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/*.a|> %B"},
		change:   []string{"dir/a.a"},
		expected: []string{"a"},
		execs:    []string{"cp dir/a.a a"},
	})
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/*.a|> %e"},
		change:   []string{"dir/a.a"},
		expected: []string{"a"},
		execs:    []string{"cp dir/a.a a"},
	})
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/*.a|> %B.%e"},
		change:   []string{"dir/a.a"},
		expected: []string{"a.a"},
		execs:    []string{"cp dir/a.a a.a"},
	})
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a.*|> b.%e 2> %Bb"},
		change:   []string{"a.a"},
		expected: []string{"b.a", "ab"},
		execs:    []string{"cp a.a b.a 2> ab"},
	})
}

func TestDirectory(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|dir/a|> dir/b"},
		change:   []string{"dir/a"},
		expected: []string{"dir/b"},
		execs:    []string{"cp dir/a dir/b"},
	})
}

func TestMultipleMatches(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a|> <|b|> dir"},
		change:   []string{"dir/", "a", "b"},
		expected: []string{"dir/a", "dir/b"},
		execs:    []string{"cp a b dir"},
	})
}

// When there are 2 marked patterns and
// one of them is a non existing file
// the rule will still be executed
func TestMultiplePatSingleMatch(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp <|a|> <|b|> dir"},
		change:   []string{"dir/", "b"},
		expected: []string{"dir/b"},
		execs:    []string{"cp a b dir"},
	})
}

// TODO: Detect infinite loops
// func TestInfiniteLoop(t *testing.T) {
// 	testEndToEnd(t, endToEndTest{
// 		madefile: []string{"cp <|a|> b","cp <|b|> a"},
// 		change:   []string{"a"},
// 		expected: []string{""},
// 		execs:    []string{"cp a dir"},
// 	})
// }

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
		if !found {
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
		abs := filepath.Join(base, fp)
		if strings.Contains(fp, "/") {
			if fp[len(fp)-1] == '/' {
				check(os.MkdirAll(abs, 0777))
				continue
			} else {
				check(os.MkdirAll(filepath.Dir(abs), 0777))
			}
		}
		check(ioutil.WriteFile(abs, []byte(""), 0777))
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
