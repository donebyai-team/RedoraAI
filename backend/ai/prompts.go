package ai

import (
	"github.com/tmc/langchaingo/prompts"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"text/template"
)

type Prompt struct {
	Model      GPTModel `json:"model"`
	PromptTmpl string   `json:"prompt_tmpl"`
	SchemaTmpl string   `json:"schema_tmpl"`
	HumanTmpl  string   `json:"human_tmpl"`
}

func (p *Prompt) getPromptTemplate(templatePrefix string, addImageSupport bool) (prompts.ChatPromptTemplate, *template.Template, []*template.Template) {
	return p.Model.getPromptTemplate(p, templatePrefix, addImageSupport)
}

////go:embed prompt/*.gotmpl
//var promptTplFS embed.FS

//func rp(name string) string {
//	//cnt, err := promptTplFS.ReadFile("prompt/" + name)
//	//if err != nil {
//	//	panic(err)
//	//}
//	//return cli.Dedent(string(cnt))
//}

type Variable map[string]any

func (v Variable) WithVariable(key string, value any) Variable {
	v[key] = value
	return v
}

func (v Variable) MergeVariable(in Variable) Variable {
	for k, val := range in {
		v[k] = val
	}
	return v
}

func humanize(in string) string {
	return cases.Title(language.AmericanEnglish).String(strings.ReplaceAll(strings.ToLower(in), "_", " "))
}
