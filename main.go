package main

import (
	"html/template"
	"log"
	"net/http"
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
    //TODO:
    pageTempl.ExecuteTemplate(w, "AddNewMemoPage", nil)
}

func addNewMemo(w http.ResponseWriter, r *http.Request){
    w.WriteHeader(400)
}
 
func main(){
    /*TODO:
    app johon voi tallentaa: otsikon, tekstiä ja kansion.
    nämä tallennetaan sqlite tietokantaan
    */
    handler := http.NewServeMux()
    server := http.Server{
        Addr: ":42069",
        Handler: handler,
    }

    // Base pages
    handler.HandleFunc("GET /", renderDisplayPage)
    handler.HandleFunc("GET /addNewMemo", renderAddNewMemoPage)

    // files
    handler.Handle("GET /files/styles.css", http.StripPrefix("/files", http.FileServer(http.Dir("css/"))))
    //TODO: lisää sivu

    // api
    handler.HandleFunc("POST /api/addNewMemo", addNewMemo)

    log.Printf("web server started on port%s\n", server.Addr)
    log.Fatal(server.ListenAndServe())
}
