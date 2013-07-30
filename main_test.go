package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	tmp, stop := TempDir()
	defer stop()

	initial := []string{"aa", "ba", "cb", "db", "Madefile"}
	createAll(tmp, initial)

	err := ioutil.WriteFile(tmp+"/Madefile", []byte("cp *a %c"), 0777)
	check(err)

	excs, err := Made(tmp)
	if err != nil {
		t.Fatal(err)
	}
	_ = excs

	built, _ := filesetDiff(listAll(tmp), initial)
	compareFilesets(t, built, []string{"aac", "bac"})
	// if len(missing) > 0 {
	// 	t.Fatal("Missing files")
	// }

	createAll(tmp, []string{"fa"})
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
			t.Error("Not in right:", l)
		}
	}
	for _, r := range right {
		if r != "" {
			t.Fail()
			t.Error("Not in left:", r)
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
