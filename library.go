package pavarotti

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const DatabaseVersion = 1

type Library struct {
	DatabasePath string
}

func (library Library) AddSong(song Song) (err error) {
	var (
		db *sql.DB
	)

	db, err = library.openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO songs (path, title, artist, album, track) VALUES (?, ?, ?, ?, ?)",
		song.Path, song.Title, song.Artist, song.Album, song.Track)

	return
}

func (library Library) openDatabase() (db *sql.DB, err error) {
	var (
		rows    *sql.Rows
		version uint
	)

	db, err = sql.Open("sqlite3", library.DatabasePath)
	if err != nil {
		return
	}

	rows, err = db.Query("SELECT version FROM schema_info")
	if err == nil && rows.Next() {
		err = rows.Scan(&version)
		if err != nil {
			db.Close()
			return
		}
	}

	for version < DatabaseVersion {
		switch version {
		case 0:
			_, err = db.Exec("CREATE TABLE songs (path TEXT, title TEXT, artist TEXT, album TEXT, track INTEGER)")
			if err != nil {
				db.Close()
				return
			}

			_, err = db.Exec("CREATE TABLE schema_info (version INTEGER)")
			if err != nil {
				db.Close()
				return
			}

			_, err = db.Exec("INSERT INTO schema_info (version) VALUES (1)")
			if err != nil {
				db.Close()
				return
			}
		}
		version += 1
	}

	return
}
