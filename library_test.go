package pavarotti

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"testing"
)

var addSongTests = []struct {
	path, title, artist, album string
	track                      uint
	err                        error
}{
	{"foo.mp3", "Foo", "Bar", "Baz", 1, nil},
}

func TestLibrary_AddSong(t *testing.T) {
	var (
		f                          *os.File
		err                        error
		db                         *sql.DB
		rows                       *sql.Rows
		path, title, artist, album string
		track                      uint
		numRows                    uint
		library                    *Library
		song                       Song
	)
	for testNum, tt := range addSongTests {
		f, err = ioutil.TempFile("", "pavarotti")
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		f.Close()
		defer os.Remove(f.Name())

		library, err = NewLibrary(f.Name())
		if err != nil {
			t.Errorf("test %d: %q", err)
			continue
		}
		song = Song{Path: tt.path, Title: tt.title, Artist: tt.artist, Album: tt.album, Track: tt.track}
		library.AddSong(song)
		err = library.Close()
		if err != tt.err {
			t.Errorf("test %d: expected %q error, got %q", testNum, tt.err, err)
			continue
		}

		db, err = sql.Open("sqlite3", f.Name())
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		defer db.Close()

		rows, err = db.Query("SELECT path, title, artist, album, track FROM songs")
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		defer rows.Close()

		numRows = 0
		for rows.Next() {
			err = rows.Scan(&path, &title, &artist, &album, &track)
			if err != nil {
				t.Errorf("test %d: %s", testNum, err)
				continue
			}

			if tt.path != path {
				t.Errorf("test %d: expected %q, got %q", tt.path, path)
			}
			if tt.title != title {
				t.Errorf("test %d: expected %q, got %q", tt.title, title)
			}
			if tt.artist != artist {
				t.Errorf("test %d: expected %q, got %q", tt.artist, artist)
			}
			if tt.album != album {
				t.Errorf("test %d: expected %q, got %q", tt.album, album)
			}
			if tt.track != track {
				t.Errorf("test %d: expected %d, got %d", tt.track, track)
			}

			numRows += 1
		}

		if numRows != 1 {
			t.Errorf("test %d: expected 1 row, got %d", testNum, numRows)
		}
	}
}
