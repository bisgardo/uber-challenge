package tpl

import (
	"strings"
	"html/template"
)

func compile(name string, funcMap template.FuncMap) *template.Template {
	name += ".tpl"
	filename := "res/tpl/" + name
	return template.Must(template.New(name).Funcs(funcMap).ParseFiles(filename))
}

var Ping = compile("ping", template.FuncMap{})

var Movie = compile("movie", template.FuncMap{
	"field": func (field string) template.HTML {
		if field == "" || field == "N/A" {
			return "<i>N/A</i>"
		}
		return template.HTML(field)
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
