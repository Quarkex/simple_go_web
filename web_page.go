package main

import (
    "html/template"
    "io/ioutil"
    "net/http"
    "regexp"
    "strings"
    "errors"
    "log"
)

var templates = template.Must(template.ParseGlob("templates/*"))
var validPath = regexp.MustCompile("^/(edit|save|view|static)/([a-zA-Z0-9]+)$")

type Page struct {
    Title string
    Body  template.HTML
}

func (p *Page) save() error {
    filename := "pages/" + p.Title + ".txt"
    return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
    filename := "pages/" + title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: template.HTML(body)}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    if r.URL.Path == "/" {
        return "index", nil
    } else if len( strings.Split(r.URL.Path, "/") ) == 2 {
        m := strings.Split(r.URL.Path, "/")
        return m[1], nil
    } else {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return "", errors.New("Invalid Page Title")
        }
        return m[2], nil // The title is the second subexpression.
    }
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        title, err := getTitle(w, r)
        if err != nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, title)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/404", http.StatusNotFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: template.HTML(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
    fs := http.FileServer(http.Dir("./static"))

    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", makeHandler(viewHandler))
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
