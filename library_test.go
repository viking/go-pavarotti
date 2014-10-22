package pavarotti

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var addSongTests = []struct {
	song Song
	err  error
}{
	{
		Song{Path: "foo.mp3", Artist: "Bar", AlbumArtist: "Qux", Album: "Baz",
			Title: "Foo", Year: 2014, Track: 1, Genre: "Stuff", Comments: []string{"Foo", "Bar"},
			Composer: "Dude", Copyright: "Bro"},
		nil,
	},
}

func TestLibrary_AddSong(t *testing.T) {
	var (
		f            *os.File
		err          error
		db           *sql.DB
		rows1, rows2 *sql.Rows
		song         Song
		library      *Library
		id           int
		comment      string
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
		library.AddSong(tt.song)
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

		// grab main song data
		rows1, err = db.Query("SELECT id, path, title, artist, albumartist, album, track, year, genre, composer, copyright FROM songs")
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		defer rows1.Close()

		if !rows1.Next() {
			t.Errorf("test %d: no rows found", testNum)
			continue
		}

		err = rows1.Scan(&id, &song.Path, &song.Title, &song.Artist, &song.AlbumArtist, &song.Album, &song.Track, &song.Year, &song.Genre, &song.Composer, &song.Copyright)
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}

		// grab song comments
		rows2, err = db.Query("SELECT data FROM comments WHERE song_id = ?", id)
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		defer rows2.Close()

		for rows2.Next() {
			err = rows2.Scan(&comment)
			if err != nil {
				t.Errorf("test %d: %s", testNum, err)
				break
			}

			song.Comments = append(song.Comments, comment)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(tt.song, song) {
			t.Errorf("test %d: expected %+v, got %+v", testNum, tt.song, song)
		}
	}
}
