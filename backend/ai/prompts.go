package ai

import (
	"embed"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/cli"
	"github.com/tmc/langchaingo/prompts"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"text/template"
	"time"
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

//go:embed prompts/*.gotmpl
var promptTplFS embed.FS

func rp(name string) string {
	cnt, err := promptTplFS.ReadFile("prompts/" + name)
	if err != nil {
		panic(err)
	}
	return cli.Dedent(string(cnt))
}

type Variable map[string]any

func (v Variable) WithCustomer(customer *models.Customer) Variable {
	v["firstName"] = customer.FirstName
	v["lastName"] = customer.LastName
	v["phoneNumber"] = customer.Phone
	return v
}

func (v Variable) WithCustomerCase(customer *models.CustomerCase) Variable {
	v["dueDate"] = customer.DueDate.Format(time.RFC3339)
	return v
}

func (v Variable) WithConversationDate(date time.Time) Variable {
	v["CalledAt"] = date.Format(time.RFC3339)
	return v
}

func (v Variable) WithCallMessages(messages []models.CallMessage) Variable {
	var atts []map[string]any
	for _, conversation := range messages {
		if conversation.SystemMessage != nil {
			continue
		}

		if conversation.UserMessage != nil {
			atts = append(atts, map[string]any{
				"Role":    "user",
				"Message": conversation.UserMessage.Message,
			})
		} else if conversation.BotMessage != nil {
			atts = append(atts, map[string]any{
				"Role":    "assistant",
				"Message": conversation.BotMessage.Message,
			})
		}
	}
	v["CallMessages"] = atts
	return v
}

func (v Variable) WithPastConversations(conversations []*models.Conversation) Variable {
	atts := make([]map[string]any, 0, len(conversations))
	for _, conversation := range conversations {
		atts = append(atts, map[string]any{
			"Date":    conversation.CreatedAt.Format(time.RFC3339),
			"Summary": conversation.Summary,
		})
	}
	v["Conversations"] = atts

	return v
}

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
