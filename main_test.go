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
	existing []string
	change   []string
	changed  []string
}

func testEndToEnd(t *testing.T, tcase endToEndTest) {
	tmp, stop := TempDir()
	defer stop()

	existing := append(tcase.existing, "Madefile")
	createAll(tmp, existing)

	err := ioutil.WriteFile(
		tmp+"/Madefile",
		[]byte(strings.Join(tcase.madefile, "\n")),
		0777)
	check(err)

	excs, err := Made(tmp)
	if err != nil {
		t.Fatal(err)
	}
	_ = excs

	// createAll(tmp, tcase.change)

	// excs2, err := Made(tmp)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// _ = excs2

	built, missing := filesetDiff(listAll(tmp), existing)
	if len(missing) > 0 {
		for _, m := range missing {
			t.Fail()
			t.Error("Made deleted:", m)
		}
	}
	compareFilesets(t, built, tcase.changed)

	for _, ex := range excs {
		t.Log(ex.String())
	}
}

func TestSimpleCommand(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp a b"},
		existing: []string{"a", "c"},
		change:   []string{},
		changed:  []string{"b"},
	})
}

func TestMultipleCommands(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp a b", "cp c d"},
		existing: []string{"a", "c"},
		change:   []string{},
		changed:  []string{"b", "d"},
	})
}

// func TestDependendCommands(t *testing.T) {
// 	testEndToEnd(t, endToEndTest{
// 		madefile: []string{"cp a b", "cp b c"},
// 		existing: []string{"a", "d"},
// 		change:   []string{},
// 		changed:  []string{"b", "c"},
// 	})
// }

func TestClassicPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp a* b"},
		existing: []string{"a", "d"},
		change:   []string{},
		changed:  []string{"b"},
	})
}

func TestForEachPattern(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp *.a %b"},
		existing: []string{"a.a", "d"},
		change:   []string{},
		changed:  []string{"a.ab"},
	})
}

func TestDoNothing(t *testing.T) {
	testEndToEnd(t, endToEndTest{
		madefile: []string{"cp c d"},
		existing: []string{"a", "b"},
		change:   []string{},
		changed:  []string{},
	})
}

func TestParseMadefile(t *testing.T) {

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
