package pavarotti

import (
	"github.com/viking/id3-go"
	"github.com/viking/id3-go/v1"
	"github.com/viking/id3-go/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type songInfo struct {
	artist      string
	albumArtist string
	album       string
	title       string
	year        int
	track       uint
	genre       string
	comments    []string
	copyright   string
	composer    string
}

var findTests = []struct {
	path     string
	tagMajor uint8
	tagMinor uint8
	tagData  songInfo
	expected songInfo
}{
	{
		// file with no tags but with good path
		path:     filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		expected: songInfo{"Foo", "", "Bar", "Baz", 0, 1, "", nil, "", ""},
	},
	{
		// file with v1 tags and good dirname
		path:     filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		tagMajor: 1,
		tagData:  songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux"},
		expected: songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 1},
	},
	{
		// file with v2.3 tags and bad path
		path:     filepath.Join("Foo", "Bar", "05 - Baz.mp3"),
		tagMajor: 2,
		tagMinor: 3,
		tagData:  songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
		expected: songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
	},
	{
		// file with v2.3 tags and missing album dir
		path:     filepath.Join("Foo", "05 - Baz.mp3"),
		tagMajor: 2,
		tagMinor: 3,
		tagData:  songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
		expected: songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
	},
	{
		// file with v2.3 tags and no dir
		path:     filepath.Join("05 - Baz.mp3"),
		tagMajor: 2,
		tagMinor: 3,
		tagData:  songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
		expected: songInfo{artist: "Foo dude", album: "Bar fight", title: "Baz qux", track: 5},
	},
	{
		// file with full v2.3 tags
		path:     filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		tagMajor: 2,
		tagMinor: 3,
		tagData:  songInfo{"Foo", "Foo", "Bar", "Baz qux", 2014, 1, "Country", []string{"Blah"}, "2014 Foo Bar", "Corge"},
		expected: songInfo{"Foo", "Foo", "Bar", "Baz qux", 2014, 1, "Country", []string{"eng\tComment:\nBlah"}, "2014 Foo Bar", "Corge"},
	},
	{
		// file with full v2.2 tags
		path:     filepath.Join("Foo", "Bar", "01 - Baz.mp3"),
		tagMajor: 2,
		tagMinor: 2,
		tagData:  songInfo{"Foo", "Foo", "Bar", "Baz qux", 2014, 1, "Country", []string{"Blah"}, "2014 Foo Bar", "Corge"},
		expected: songInfo{"Foo", "Foo", "Bar", "Baz qux", 2014, 1, "Country", []string{"eng\tComment:\nBlah"}, "2014 Foo Bar", "Corge"},
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
		if tt.tagMajor > 0 {
			var tagger id3.Tagger

			if tt.tagMajor == 1 {
				tagger = new(v1.Tag)
			} else if tt.tagMajor == 2 {
				tagger = v2.NewTag(tt.tagMinor)

				var (
					ft         v2.FrameType
					textFrame  *v2.TextFrame
					utextFrame *v2.UnsynchTextFrame
				)

				// add track
				if tt.tagMinor == 2 {
					ft = v2.V22FrameTypeMap["TRK"]
				} else if tt.tagMinor == 3 {
					ft = v2.V23FrameTypeMap["TRCK"]
				}
				textFrame = v2.NewTextFrame(ft, strconv.FormatUint(uint64(tt.tagData.track), 10))
				tagger.AddFrames(textFrame)

				// add album artist
				if tt.tagMinor == 2 {
					ft = v2.V22FrameTypeMap["TP2"]
				} else if tt.tagMinor == 3 {
					ft = v2.V23FrameTypeMap["TPE2"]
				}
				textFrame = v2.NewTextFrame(ft, tt.tagData.albumArtist)
				tagger.AddFrames(textFrame)

				// add comments
				for _, comment := range tt.tagData.comments {
					if tt.tagMinor == 2 {
						ft = v2.V22FrameTypeMap["COM"]
					} else if tt.tagMinor == 3 {
						ft = v2.V23FrameTypeMap["COMM"]
					}
					utextFrame = v2.NewUnsynchTextFrame(ft, "Comment", comment)
					tagger.AddFrames(utextFrame)
				}

				// add copyright
				if tt.tagMinor == 2 {
					ft = v2.V22FrameTypeMap["TCR"]
				} else if tt.tagMinor == 3 {
					ft = v2.V23FrameTypeMap["TCOP"]
				}
				textFrame = v2.NewTextFrame(ft, tt.tagData.copyright)
				tagger.AddFrames(textFrame)

				// add composer
				if tt.tagMinor == 2 {
					ft = v2.V22FrameTypeMap["TCM"]
				} else if tt.tagMinor == 3 {
					ft = v2.V23FrameTypeMap["TCOM"]
				}
				textFrame = v2.NewTextFrame(ft, tt.tagData.composer)
				tagger.AddFrames(textFrame)
			}

			tagger.SetTitle(tt.tagData.title)
			tagger.SetArtist(tt.tagData.artist)
			tagger.SetAlbum(tt.tagData.album)
			tagger.SetYear(strconv.Itoa(tt.tagData.year))
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
					t.Errorf("test %d: expected year to be %d, got %d", testNum, tt.expected.year, song.Year)
				}
				if song.Track != tt.expected.track {
					t.Errorf("test %d: expected track to be %d, got %d", testNum, tt.expected.track, song.Track)
				}
				if song.Genre != tt.expected.genre {
					t.Errorf("test %d: expected genre to be <%q>, got <%q>", testNum, tt.expected.genre, song.Genre)
				}
				if len(tt.expected.comments) != len(song.Comments) {
					t.Errorf("test %d: expected comments to be %q, got %q", testNum, tt.expected.comments, song.Comments)
				} else {
					for i, comment := range tt.expected.comments {
						if comment != song.Comments[i] {
							t.Errorf("test %d: expected comments to be %q, got %q", testNum, tt.expected.comments, song.Comments)
							break
						}
					}
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
