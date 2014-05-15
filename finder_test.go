package pavarotti

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFind_PathParsing(t *testing.T) {
	var (
		dir string
		err error
		f   *os.File
	)
	dir, err = ioutil.TempDir("", "pavarotti")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dir)
	defer os.RemoveAll(dir)

	albumPath := filepath.Join(dir, "Foo", "Bar")
	err = os.MkdirAll(albumPath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	songPath := filepath.Join(albumPath, "01 - Baz.mp3")
	f, err = os.Create(songPath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	ch := make(chan Song)
	quit := make(chan bool)
	go func() {
		var song Song
		select {
		case song = <-ch:
			if song.Path != songPath {
				t.Errorf("expected %v, got %v", songPath, song.Path)
			}
		case <-quit:
			t.Error("no song found")
			return
		}

		select {
		case <-ch:
			t.Error("extra song found")
		case <-quit:
		}
	}()

	err = Find(dir, ch)
	if err != nil {
		t.Error(err)
	}
	quit <- true
}
