package tpl

import (
	"src/config"
	"src/logging"
	"strings"
	"html/template"
	"net/http"
	"reflect"
	"appengine"
)

const extension = ".tpl"

type Data []interface{}

func compile(name string, funcMap template.FuncMap) *template.Template {
	tplName := name + extension
	filename := "res/tpl/" + tplName
	
	// Writing comments with functions is needed to prevent engine from filtering them out.
	funcMap["comment_begin"] = func () template.HTML { return "<!--" }
	funcMap["comment_end"] = func () template.HTML { return "-->" }
	funcMap["has_field"] = func (arg interface{}, field string) bool {
		v := reflect.Indirect(reflect.ValueOf(arg))
		_, exists := v.Type().FieldByName(field)
		return exists
	}
	funcMap["maps_api_key"] = func () string {
		return config.MapsApiKey()
	}
	
	return template.Must(template.New(tplName).Funcs(funcMap).ParseFiles("res/tpl/layout.tpl", filename))
}

type TemplateData struct {
	Subtitle string
	Version  string
	Log      []string
	Data     interface{}
}

func NewTemplateData(ctx appengine.Context, logger *logging.RecordingLogger, data interface{}) TemplateData {
	version := appengine.VersionID(ctx)
	log := logger.Entries
	return TemplateData{Version: version, Log: log, Data: data}
}

func Render(w http.ResponseWriter, tpl *template.Template, data TemplateData) error {
	return tpl.ExecuteTemplate(w, "layout", data)
}

var About = compile("about", template.FuncMap{})

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
