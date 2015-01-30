package main

//protocol generator, a file a generated around each hittype

import (
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/huandu/xstrings"
)

const protocolV1 = "https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters"

type HitType struct {
	Name,
	StructName string
	Fields     []*Field
	Indices    []string
	Required   bool
	HitTypeIDs []string
}

type Field struct {
	Name,
	Docs,
	Param,
	ParamStr,
	Type,
	Default,
	MaxLen,
	Examples string
	Required bool
	HitTypes []string
	Indices  []string
}

//===================================

var allFields = []*Field{}

var hitTypes = map[string]*HitType{
	//special base type
	"client": &HitType{Name: "client"},
}

//====================================================
//the code template (split into separate strings for somewhat better readability)

func buildCode() string {

	var clientFields = `{{if eq .Name "client" }}	//Use TLS when Send()ing
	UseTLS bool
{{end}}`

	//@TODO {{.Type}}
	var fields = `{{range $index, $element := .Fields}}{{.Docs}}
	{{.Name}} string
{{end}}`

	var indices = `{{range $index, $value := .Indices}}	// {{$value}} is required by other properties
	{{$value}} string
{{end}}`

	var paramRequired = `{{if .Required}} else {
		return errors.New("Required field '{{.Name}}' is missing")
	}{{end}}`

	var params = `{{range $index, $element := .Fields}}	if h.{{.Name}} != "" {
		v.Add({{.ParamStr}}, h.{{.Name}})
	}` + paramRequired + `
{{end}}`

	var clientFuncs = `{{if eq .Name "client" }}
func (c *Client) setType(h hitType) {
	switch h.(type) {
{{range $index, $id := .HitTypeIDs}}	case *{{ exportName $id}}:
		c.hitType = "{{ $id }}"
{{end}}	}
}
{{end}}`

	return `package ga

//WARNING: This file was generated. Do not edit.

import "net/url"{{if .Required}}
import "errors"{{end}}
` + clientFuncs + `
//{{.StructName}} Hit Type
type {{.StructName}} struct {
` + clientFields +
		fields +
		indices +
		`}

func (h *{{.StructName}}) addFields(v url.Values) error {
` + params + `	return nil
}`
}

var codeTemplate *template.Template

//====================================================
//helper functions

var trim = strings.TrimSpace

func commentIndent(s string) string {
	words := strings.Split(s, " ")
	comment := "\t//"
	width := 0
	for _, w := range words {
		if width > 55 {
			width = 0
			comment += "\n\t//"
		}
		width += len(w) + 1
		comment += " " + w
	}
	return comment
}

func exportName(s string) string {
	words := strings.Split(s, " ")
	for i, w := range words {
		words[i] = xstrings.FirstRuneToUpper(w)
	}
	return strings.Join(words, "")
}

var preslash = grepper(`^([^\/]+)`)
var alpha = striper(`[^A-Za-z]`)

func paramName(s string) string {
	return alpha(exportName(preslash(s)))
}

func goName(parent, prop string) string {
	return strings.TrimPrefix(prop, parent)
}

func goType(gaType string) string {
	switch gaType {
	case "text":
		return "string"
	case "integer":
		return "int64"
	case "boolean":
		return "bool"
	case "currency":
		return "float64"
	}
	log.Fatal("Unknown GA Type: " + gaType)
	return ""
}

var templateFuncs = template.FuncMap{
	"exportName": exportName,
}

var indexMatcher = regexp.MustCompile(`<[A-Za-z]+>`)
var indexVar = grepper(`<([A-Za-z]+)>`)
var strVar = grepper(`'([a-z]+)'`)

//meta helpers
func grepper(restr string) func(string) string {
	re := regexp.MustCompile(restr)
	return func(s string) string {
		matches := re.FindAllStringSubmatch(s, 1)
		if len(matches) != 1 {
			log.Fatalf("'%s' should match '%s' exactly once", restr, s)
		}
		groups := matches[0]
		if len(groups) != 2 {
			log.Fatalf("'%s' should have exactly one group (found %d)", restr, len(groups))
		}
		return groups[1]
	}
}

func striper(restr string) func(string) string {
	re := regexp.MustCompile(restr)
	return func(s string) string {
		return re.ReplaceAllString(s, "")
	}
}

//====================================

//main pipeline
func main() {
	check()
}

func check() {
	code := buildCode()
	t := template.New("ga-code")
	t = t.Funcs(templateFuncs)
	t, err := t.Parse(code)
	if err != nil {
		log.Fatalf("Template error: %s", err)
	}
	codeTemplate = t
	parse()
}

func parse() {
	doc, err := goquery.NewDocument(protocolV1)
	if err != nil {
		log.Fatal(err)
	}

	//special exluded field
	var hitTypeDocs string

	doc.Find("h3").Each(func(i int, s *goquery.Selection) {

		content := s.Next().Children()
		cells := content.Eq(2).Find("tr td")

		//get trimmed raw contents
		f := &Field{
			Name:     trim(s.Find("a").Text()),
			Required: !strings.Contains(content.Eq(0).Text(), "Optional"),
			Docs:     trim(content.Eq(1).Text()),
			Param:    trim(cells.Eq(0).Text()),
			Type:     trim(cells.Eq(1).Text()),
			Default:  trim(cells.Eq(2).Text()),
			MaxLen:   trim(cells.Eq(3).Text()),
			HitTypes: strings.Split(trim(cells.Eq(4).Text()), ", "),
			Examples: trim(content.Eq(3).Text()),
		}

		if f.Name == "Hit type" {
			hitTypeDocs = f.Docs
		}
		allFields = append(allFields, f)
	})

	buildTypes(hitTypeDocs)
}

func buildTypes(hitTypeDocs string) {
	//hit type docs contain all type names
	for _, t := range strings.Split(hitTypeDocs, ",") {
		t = strVar(t)
		hitTypes[t] = &HitType{Name: t}
	}

	//place each field in one or more types
	for _, f := range allFields {
		for _, t := range f.HitTypes {
			//the client type holds all the common fields
			if t == "all" {
				t = "client"
			}
			h, exists := hitTypes[t]
			if !exists {
				log.Fatalf("Unknown type: '%s'", t)
			}
			h.Fields = append(h.Fields, f)
		}
	}

	process()
}

func process() {
	for _, f := range allFields {
		processField(f)
	}
	for _, h := range hitTypes {
		processHitType(h)
	}
	generate()
}

func processField(f *Field) {

	f.Name = paramName(f.Name)

	//unexport the hit type field
	if f.Name == "HitType" {
		f.Name = xstrings.FirstRuneToLower(f.Name)
	}

	f.Docs = commentIndent(trim(f.Docs))
	f.Type = goType(f.Type)
	f.ParamStr = `"` + f.Param + `"`

	//check param for <extraVars>
	for _, i := range indexMatcher.FindAllString(f.Param, -1) {
		//extract each var
		newi := xstrings.FirstRuneToUpper(indexVar(i))
		f.Indices = append(f.Indices, newi)

		//convert param var into a string concat
		f.ParamStr = strings.Replace(f.ParamStr, i, `" + h.`+newi+` + "`, 1)
	}
}

func processHitType(h *HitType) {

	h.StructName = exportName(h.Name)

	//set of index vars
	is := map[string]bool{}
	for _, f := range h.Fields {
		if h.Name != "client" {
			//trim the field name by the hittype name - prevents ga.Event.EventAction
			f.Name = goName(h.StructName, f.Name)
		}
		//add all index vars to the set
		for _, i := range f.Indices {
			is[i] = true
		}

		//this type has a required field
		if f.Required {
			h.Required = true
		}
	}
	//extra psuedo-fields to be added
	for i, _ := range is {
		h.Indices = append(h.Indices, i)
	}

	//place all non-client hittype ids in client for templating
	if h.Name == "client" {
		for _, h2 := range hitTypes {
			if h2.Name != "client" {
				h.HitTypeIDs = append(h.HitTypeIDs, h2.Name)
			}
		}
		sort.Strings(h.HitTypeIDs)
	}
}

func generate() {
	for _, h := range hitTypes {
		generateFile(h)
	}
}

func generateFile(h *HitType) {
	f, err := os.Create("type-" + h.Name + ".go")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	codeTemplate.Execute(f, h)
}
