package web

import (
	"html/template"
	"net/http"
)

var T *template.Template

func ParseTemplates(dir string, fnc template.FuncMap) {
	T = template.Must(template.New("main").Funcs(fnc).ParseGlob(dir + "/*"))
}

func Execute(w http.ResponseWriter, name string, data interface{}) error {
	err := T.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}

	return nil
}
