package main

import (
    "os"
    "path/filepath"
    "html/template"
    "io/ioutil"
    "net/http"
    "strings"
    "errors"
    "log"
)

var templates = template.Must(template.ParseGlob("templates/*"))

type Page struct {
    Title string
    Body  template.HTML
}

func (p *Page) save() error {
    filename := "pages/" + p.Title + ".htm"
    dirname := filepath.Dir(filename)

    err := os.MkdirAll(dirname,0700)
    if err != nil {
        return err
    }

    return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
    filename := "pages/" + title + ".htm"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: template.HTML(body)}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    if r.URL.Path == "/" {
        return "index", nil
    } else {
        m := strings.Split(r.URL.Path, "/")
        path := []string {}
        for i, v := range m {
            if (i == 0) {
                continue
            } else if (i == 1) {
                if v == "edit" || v == "save" || v == "view" || v == "static" {
                    continue
                }
            }
            path = append(path, v)
            if v == ".." {
                http.NotFound(w, r)
                return "", errors.New("Invalid Page Title")
            }
        }
        return strings.Join(path[:],"/"), nil
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
    if title == "index" {
        http.Redirect(w, r, "/", http.StatusFound)
    } else {
        http.Redirect(w, r, "/"+title, http.StatusFound)
    }
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
