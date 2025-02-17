package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// contains the templates for responding to requests
var respTempl *template.Template
// contains the full page views
var pageTempl *template.Template

func init(){
    var err error

    respTempl, err = template.ParseGlob("./views/*.html")
    if err != nil{
        log.Fatalf("Error parsing response templates %s\n", err)
    }

    pageTempl, err = template.ParseGlob("./pages/*.html")
    if err != nil {
        log.Fatalf("Error parsing page templates %s\n", err)
    }
}

func renderDisplayPage(w http.ResponseWriter, r *http.Request){
    pageTempl.ExecuteTemplate(w, "DisplayPage", nil)
}

func renderAddNewMemoPage(w http.ResponseWriter, r *http.Request){
    pageTempl.ExecuteTemplate(w, "AddNewMemoPage", nil)
}

func addNewMemo(db *sql.DB, w http.ResponseWriter, r *http.Request){
    //TODO: tarkista ettei input sisällä ei sallittuja merkkejä
    r.ParseForm()
    insertNewNoteQuery := `INSERT INTO notes(title, note) VALUES(?,?);`
    insertNewTagQuery := `INSERT INTO tags(tag_name) VALUES(?);`
    insertNewLinkQuery := `INSERT INTO note_tag_link(note_id, tag_id) VALUES(
        (SELECT id FROM notes WHERE note = ?), (SELECT id FROM tags where tag_name = ?)
    );`

    title := r.FormValue("title")
    note := r.FormValue("muistiinpano")
    tags := strings.Split(r.FormValue("tags"), " ")

    _, err := db.Exec(insertNewNoteQuery, title, note)
    if err != nil{
        w.WriteHeader(500)
        log.Printf("error adding new note to the database: %s\n", err)
        return
    }

    for _, tag := range tags{
        _, err := db.Exec(insertNewTagQuery, tag)
        if err != nil && err.Error() != "UNIQUE constraint failed: tags.tag_name"{
            w.WriteHeader(500)
            log.Printf("error adding tags to the database: %s\n", err)
            return
        }
        _, err = db.Exec(insertNewLinkQuery, note, tag)
        if err != nil {
            w.WriteHeader(500)
            log.Printf("error linking tags and note: %s\n", err)
            return
        }
    }

    w.WriteHeader(200)
}

func connectDb() (*sql.DB, error){
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil{
        return nil, err
    }
    //NOTE: eikö tää error logiikka tässä tapauksessa oo vähän overkill?
    // eikös sitä vois vaa returnaa db ja error?

    err = db.Ping()
    if err != nil{
        return nil, err
    }
    return db, nil
}

func initalizeDb(db *sql.DB) error{
    query := `CREATE TABLE IF NOT EXISTS notes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        note TEXT NOT NULL UNIQUE
    );
    CREATE TABLE IF NOT EXISTS tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        tag_name TEXT NOT NULL UNIQUE
    );
    CREATE TABLE IF NOT EXISTS note_tag_link (
        note_id INTEGER NOT NULL,
        tag_id INTEGER NOT NULL,
        PRIMARY KEY (note_id, tag_id),
        FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
    );`

    _, err := db.Exec(query)
    return err
}

type note struct{
    Id int
    Title string
    Text string
    Tags []string
}

func getAllNotes(db *sql.DB) ([]note, error){
    fetchNotes := `SELECT id, title, note FROM notes;`
    fetchTags := `select t.tag_name from tags as t
    inner join note_tag_link as ntl
    on t.id = ntl.tag_id
    inner join notes as n
    on ntl.note_id = n.id
    WHERE n.id = ?; `

    var notes []note

    dbNotes, err := db.Query(fetchNotes)
    if err != nil{
        log.Printf("Error fetching all notes from the database %s\n", err)
        return nil, err
    }

    for dbNotes.Next(){
        var note note
        var id int
        dbNotes.Scan(&id, &note.Title, &note.Text)

        dbTags, err := db.Query(fetchTags, id)
        if err != nil {
            log.Printf("Error fetching tags for displaying all notes: %s\n", err)
            return nil, err
        }

        for dbTags.Next(){
            var scannedTag string
            dbTags.Scan(&scannedTag)
            note.Tags = append(note.Tags, scannedTag)
        }

        notes = append(notes, note)
        dbTags.Close()
    }
    dbNotes.Close()

    return notes, nil
}

func main(){
    handler := http.NewServeMux()
    server := http.Server{
        Addr: ":42069",
        Handler: handler,
    }

    db, err := connectDb()
    if err != nil{
        log.Fatalf("Error connecting to the database: %s\n", err)
    }
    err = initalizeDb(db)
    if err != nil{
        log.Fatalf("Error initializing the database schema: %s\n", err)
    }
    defer db.Close()

    // Base pages
    handler.HandleFunc("GET /", renderDisplayPage)
    handler.HandleFunc("GET /addNewMemo", renderAddNewMemoPage)

    // files
    handler.Handle("GET /files/styles.css", http.StripPrefix("/files", http.FileServer(http.Dir("css/"))))

    // api
    handler.HandleFunc("POST /api/addNewMemo", func(w http.ResponseWriter, r *http.Request) {
        addNewMemo(db, w, r)
    })

    //NOTE: testi
    handler.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
        getAllNotes(db)
    })

    log.Printf("Note taking web app server started on port%s\n", server.Addr)
    log.Fatal(server.ListenAndServe())
}
