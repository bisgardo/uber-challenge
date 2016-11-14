package tpl

import (
	"strings"
	"html/template"
	"net/http"
	"reflect"
	"log"
)

const extension = ".tpl"

type Data []interface{}

func compile(name string, funcMap template.FuncMap) *template.Template {
	tplName := name + extension
	filename := "res/tpl/" + tplName
	
	funcMap["logs_comment_begin"] = func () template.HTML { return "<!-- <LOGS>" }
	funcMap["logs_comment_end"] = func () template.HTML { return "</LOGS> -->" }
	funcMap["has_field"] = func (arg interface{}, field string) bool {
		v := reflect.Indirect(reflect.ValueOf(arg))
		f, exists := v.Type().FieldByName(field)
		log.Printf("%v has field '%s': %b (%s)\n", arg, field, exists, f)
		return exists
	}
	
	return template.Must(template.New(tplName).Funcs(funcMap).ParseFiles("res/tpl/layout.tpl", filename))
}

func Render(w http.ResponseWriter, tpl *template.Template, args interface{}) error {
	return tpl.ExecuteTemplate(w, "layout", args)
}

var Front = compile("front", template.FuncMap{})

var Movie = compile("movie", template.FuncMap{
	"field": func (field string) template.HTML {
		if field == "" || field == "N/A" {
			return "<i>N/A</i>"
		}
		// Escape manually.
		return template.HTML(template.HTMLEscapeString(field))
	},
})

var Movies = compile("movies", template.FuncMap{
	"join": func(ss []string) string {
		switch len(ss) {
		case 0:
			return ""
		case 1:
			return ss[0]
		case 2:
			return ss[0] + " and " + ss[1]
		}
		return strings.Join(ss[:len(ss) - 1], ", ") + ", and " + ss[len(ss) - 1]
	},
	"parenthesize": func(s string) string {
		return "(" + s + ")"
	},
})

var Status = compile("status", template.FuncMap{})

var Ping = compile("ping", template.FuncMap{})
