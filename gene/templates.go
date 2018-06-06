package gene

import (
	"bytes"
	"encoding/json"
	"go/format"
	"html/template"
	"log"
)

var bootTpl = `package main

import "{{.rootDir}}/imports"

func init() {
	app.Init()

	// ###
}

// auto generate by _, dont modify`

var importerTpl = `package imports

import core "github.com/shaomingquan/webcore"
import "{{.rootDir}}{{.pkgDir}}"
{{range $pkg := .pkgs}}
import {{$pkg.pkgid}} "{{$.rootDir}}/{{$pkg.pkg}}"
{{end}}

func Start{{.appid}}(app *core.App) {
	{{range $item := .midwares}}
	app.MidWare("{{$.prefix}}", {{$item.pkgid}}.{{$item.method}}({{$item.params}}))
	{{end}}

	{{range $item := .routers}}
	app.Router(
		"{{$.prefix}}",
		{{$.appName}}.MethodOf{{$item}}, 
		{{$.appName}}.PrefixOf{{$item}}, 
		{{$.appName}}.HandlerOf{{$item}},
	)
	{{end}}
}

// auto generate by _, dont modify`

func tplCommon(data map[string]interface{}, tpl string) []byte {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}
	buff := bytes.NewBufferString("")
	err = t.Execute(buff, data)
	if err != nil {
		log.Fatal(err)
	}
	ret, err := format.Source(buff.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

// MakeBootFile boot file template
func MakeBootFile(data map[string]interface{}) []byte {
	return tplCommon(data, bootTpl)
}

// MakeImporterFile importer file template
func MakeImporterFile(data map[string]interface{}) []byte {
	databyted, _ := json.Marshal(data)
	println(string(databyted))
	return tplCommon(data, importerTpl)
}
