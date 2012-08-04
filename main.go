package main

import (
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

var (
	root string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get PWD: %v", err)
	}
	flag.StringVar(&root, "root", wd, "document root")
	flag.Parse()
}

func PageHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	page := vars["key"]
	md, err := loadMetadata(page)
	if err != nil {
		panic(err)
	}
	content, err := loadContent(page)
	if err != nil {
		panic(err)
	}
	tmpl, err := loadTemplate()
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(resp, struct {
		Metadata map[string]interface{}
		Content  string
	}{
		md,
		string(content),
	}); err != nil {
		panic(err)
	}
}

func loadMetadata(page string) (map[string]interface{}, error) {
	path := filepath.Join(root, "pages", page, "metadata.json")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	md := make(map[string]interface{})
	if err := decoder.Decode(&md); err != nil {
		return nil, err
	}
	return md, nil
}

func loadContent(page string) ([]byte, error) {
	path := filepath.Join(root, "pages", page, "content.md")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return blackfriday.MarkdownCommon(content), nil
}

func loadTemplate() (*template.Template, error) {
	path := filepath.Join(root, "pages", "template.html")
	return template.ParseFiles(path)
}

func AssetHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fmt.Fprintf(resp, "You asked for asset %v for page %v\n", vars["asset"], vars["key"])
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{key}", PageHandler)
	r.HandleFunc("/{key}/{asset}", AssetHandler)
	r.HandleFunc("/static/", http.FileServer(http.Dir(filepath.Join(root, "static"))))
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
