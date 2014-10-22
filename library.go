package pavarotti

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const DatabaseVersion = 1

type Library struct {
	DatabasePath string
	queue        []Song
	db           *sql.DB
}

func NewLibrary(databasePath string) (library *Library, err error) {
	library = &Library{DatabasePath: databasePath}
	err = library.openDatabase()
	return
}

func (library *Library) AddSong(song Song) {
	library.queue = append(library.queue, song)
	return
}

func (library *Library) Flush() (err error) {
	var (
		tx   *sql.Tx
		stmt *sql.Stmt
	)

	tx, err = library.db.Begin()
	if err != nil {
		return
	}

	stmt, err = tx.Prepare("INSERT INTO songs (path, title, artist, album, track) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return
	}

	for _, song := range library.queue {
		_, err = stmt.Exec(song.Path, song.Title, song.Artist, song.Album, song.Track)
		if err != nil {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	library.queue = nil
	return
}

func (library *Library) Close() (err error) {
	if library.db != nil {
		err = library.Flush()
		library.db.Close()
		library.db = nil
	}
	return
}

func (library *Library) openDatabase() (err error) {
	var (
		rows    *sql.Rows
		version uint
	)

	library.db, err = sql.Open("sqlite3", library.DatabasePath)
	if err != nil {
		return
	}

	rows, err = library.db.Query("SELECT version FROM schema_info")
	if err == nil && rows.Next() {
		err = rows.Scan(&version)
		if err != nil {
			library.db.Close()
			return
		}
	}

	for version < DatabaseVersion {
		switch version {
		case 0:
			_, err = library.db.Exec("CREATE TABLE songs (path TEXT, title TEXT, artist TEXT, album TEXT, track INTEGER)")
			if err != nil {
				library.db.Close()
				return
			}

			_, err = library.db.Exec("CREATE TABLE schema_info (version INTEGER)")
			if err != nil {
				library.db.Close()
				return
			}

			_, err = library.db.Exec("INSERT INTO schema_info (version) VALUES (1)")
			if err != nil {
				library.db.Close()
				return
			}
		}
		version += 1
	}

	return
}
