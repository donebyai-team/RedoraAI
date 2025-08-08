package ai

import (
	"strings"
	"time"

	"github.com/shank318/doota/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetVars(customerCase *models.AugmentedCustomerCase, currentTime time.Time) Variable {
	out := make(Variable).
		WithCustomer(customerCase.Customer).
		WithCustomerCase(customerCase.CustomerCase).
		WithPastConversations(customerCase.Conversations).
		WithConversationDate(currentTime)
	return out
}

func GetCaseDecisionVars(customerCase *models.Conversation) Variable {
	out := make(Variable).
		WithConversationDate(customerCase.CreatedAt).
		WithCallMessages(customerCase.CallMessages)
	return out
}

func GetSubRedditRulesEvalVars(subReddit *models.Source) Variable {
	out := make(Variable)
	out["Name"] = subReddit.Name
	out["Rules"] = subReddit.Metadata.Rules
	return out
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

func GetPostGenerationVars(input *PostGenerateInput) Variable {
	out := make(Variable)
	out["Topic"] = input.PostSetting.Topic
	out["Context"] = input.PostSetting.Context
	out["Goal"] = input.PostSetting.Goal
	out["Tone"] = input.PostSetting.Tone
	out["Rules"] = input.Rules
	out["Flairs"] = input.Flairs
	return out
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
