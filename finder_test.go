package pavarotti

import (
	"github.com/mikkyang/id3-go"
	"github.com/mikkyang/id3-go/v1"
	"github.com/mikkyang/id3-go/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type songInfo struct {
	artist      string
	albumArtist string
	album       string
	title       string
	year        string
	track       string
	genre       string
	comment     string
	copyright   string
	composer    string
}

var findTests = []struct {
	path       string
	tagVersion uint8
	tagData    songInfo
	expected   songInfo
}{
	{
		path:     filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		expected: songInfo{"Foo", "", "Bar", "Baz", "", "01", "", "", "", ""},
	},
	{
		path:       filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		tagVersion: 1,
		tagData:    songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux"},
		expected:   songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: "01"},
	},
	{
		path:       filepath.Join("Foo", "Bar", "05 - Baz.mp3"),
		tagVersion: 2,
		tagData:    songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: "01"},
		expected:   songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: "01"},
	},
}

func TestFind(t *testing.T) {
	var (
		dir string
		err error
		f   *os.File
	)

	for testNum, tt := range findTests {
		dir, err = ioutil.TempDir("", "pavarotti")
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		t.Logf("test %d: %s", testNum, dir)
		defer os.RemoveAll(dir)

		songPath := filepath.Join(dir, tt.path)
		songDir := filepath.Dir(songPath)
		if songDir != "" {
			err = os.MkdirAll(songDir, 0755)
			if err != nil {
				t.Errorf("test %d: %s", testNum, err)
				continue
			}
		}

		f, err = os.Create(songPath)
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
			continue
		}
		if tt.tagVersion > 0 {
			var tagger id3.Tagger

			if tt.tagVersion == 1 {
				tagger = new(v1.Tag)
			} else if tt.tagVersion == 2 {
				tagger = v2.NewTag(3)
				ft := v2.V23FrameTypeMap["TRCK"]
				textFrame := v2.NewTextFrame(ft, tt.tagData.track)
				tagger.AddFrames(textFrame)
			}

			tagger.SetTitle(tt.tagData.title)
			tagger.SetArtist(tt.tagData.artist)
			tagger.SetAlbum(tt.tagData.album)
			tagger.SetYear(tt.tagData.year)
			tagger.SetGenre(tt.tagData.genre)
			if _, err = f.Write(tagger.Bytes()); err != nil {
				f.Close()
				t.Errorf("test %d: %s", testNum, err)
				continue
			}
		}
		f.Close()

		ch := make(chan Song)
		quit := make(chan bool)
		go func() {
			var song Song
			select {
			case song = <-ch:
				if song.Path != songPath {
					t.Errorf("test %d: expected path to be <%q>, got <%q>", testNum, songPath, song.Path)
				}
				if song.Artist != tt.expected.artist {
					t.Errorf("test %d: expected artist to be <%q>, got <%q>", testNum, tt.expected.artist, song.Artist)
				}
				if song.AlbumArtist != tt.expected.albumArtist {
					t.Errorf("test %d: expected albumArtist to be <%q>, got <%q>", testNum, tt.expected.albumArtist, song.AlbumArtist)
				}
				if song.Album != tt.expected.album {
					t.Errorf("test %d: expected album to be <%q>, got <%q>", testNum, tt.expected.album, song.Album)
				}
				if song.Title != tt.expected.title {
					t.Errorf("test %d: expected title to be <%q>, got <%q>", testNum, tt.expected.title, song.Title)
				}
				if song.Year != tt.expected.year {
					t.Errorf("test %d: expected year to be <%q>, got <%q>", testNum, tt.expected.year, song.Year)
				}
				if song.Track != tt.expected.track {
					t.Errorf("test %d: expected track to be <%q>, got <%q>", testNum, tt.expected.track, song.Track)
				}
				if song.Genre != tt.expected.genre {
					t.Errorf("test %d: expected genre to be <%q>, got <%q>", testNum, tt.expected.genre, song.Genre)
				}
				if song.Comment != tt.expected.comment {
					t.Errorf("test %d: expected comment to be <%q>, got <%q>", testNum, tt.expected.comment, song.Comment)
				}
				if song.Composer != tt.expected.composer {
					t.Errorf("test %d: expected composer to be <%q>, got <%q>", testNum, tt.expected.composer, song.Composer)
				}
				if song.Copyright != tt.expected.copyright {
					t.Errorf("test %d: expected copyright to be <%q>, got <%q>", testNum, tt.expected.copyright, song.Copyright)
				}
			case <-quit:
				t.Errorf("test %d: no song found", testNum)
				return
			}

			select {
			case <-ch:
				t.Errorf("test %d: extra song found", testNum)
			case <-quit:
			}
		}()

		err = Find(dir, ch)
		if err != nil {
			t.Errorf("test %d: %s", testNum, err)
		}
		quit <- true
	}
}
