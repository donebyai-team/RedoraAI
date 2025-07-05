package pbportal

import (
	"fmt"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"github.com/shank318/doota/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

func (r *UserRole) FromModel(model models.UserRole) {
	value := "USER_ROLE_" + strings.ToUpper(model.String())
	enum, found := UserRole_value[value]
	if !found {
		panic(fmt.Errorf("unknown user role model %q", model))
	}

	*r = UserRole(enum)
}

func (r *NotificationFrequency) FromModel(model models.NotificationFrequency) {
	value := "NOTIFICATION_FREQUENCY_" + strings.ToUpper(model.String())
	enum, found := NotificationFrequency_value[value]
	if !found {
		panic(fmt.Errorf("unknown notification frequency %q", model))
	}

	*r = NotificationFrequency(enum)
}

func (c *NotificationFrequency) ToModel() models.NotificationFrequency {
	value := strings.TrimPrefix(strings.ToUpper(c.String()), "NOTIFICATION_FREQUENCY_")
	model := models.NotificationFrequency(value)
	if !model.IsValid() {
		panic(fmt.Errorf("unknown notification frequency pb %q", value))
	}

	return model
}

func (u *User) FromModel(model *models.User, orgs []*models.Organization) *User {
	u.Id = model.ID
	u.Email = model.Email
	u.EmailVerified = model.EmailVerified
	u.Role.FromModel(model.Role)
	u.Organizations = make([]*Organization, len(orgs))
	for i, org := range orgs {
		u.Organizations[i] = new(Organization).FromModel(org)
	}
	u.CreatedAt = timestamppb.New(model.CreatedAt)
	return u
}

func (o *Organization) FromModel(model *models.Organization) *Organization {
	o.Id = model.ID
	o.Name = model.Name
	o.FeatureFlags = &OrganizationFeatureFlags{}
	if model.FeatureFlags.Subscription != nil {
		o.FeatureFlags.Subscription = new(pbcore.Subscription).FromModel(model.FeatureFlags.Subscription)
		o.FeatureFlags.Subscription.MaxSources = int32(model.FeatureFlags.GetMaxSourcesAllowed())
		o.FeatureFlags.Subscription.MaxKeywords = int32(model.FeatureFlags.GetMaxKeywordAllowed())
	}
	o.FeatureFlags.Comment = &AutomationSetting{
		Enabled:        model.FeatureFlags.EnableAutoComment,
		RelevancyScore: float32(model.FeatureFlags.GetRelevancyScoreComment()),
		MaxPerDay:      model.FeatureFlags.GetMaxCommentsPerDay(),
	}

	o.FeatureFlags.DM = &AutomationSetting{
		Enabled:        model.FeatureFlags.EnableAutoDM,
		RelevancyScore: float32(model.FeatureFlags.GetRelevancyScoreDM()),
		MaxPerDay:      model.FeatureFlags.GetMaxDMsPerDay(),
	}

	o.FeatureFlags.NotificationSettings = &NotificationSettings{}
	o.FeatureFlags.NotificationSettings.RelevantPostFrequency.FromModel(model.FeatureFlags.GetNotificationFrequency())

	o.CreatedAt = timestamppb.New(model.CreatedAt)
	return o
}

// Convert models.PostMetadata to pbcore.PostMetadata
func convertMetadata(meta models.PostMetadata) *pbcore.PostMetadata {
	history := []*pbcore.PostRegenerationHistory{}
	for _, h := range meta.History {
		history = append(history, &pbcore.PostRegenerationHistory{
			Text: h.Text,
			PostSettings: &pbcore.PostSettings{
				Topic:       h.PostSettings.Topic,
				Context:     h.PostSettings.Context,
				Goal:        h.PostSettings.Goal,
				Tone:        h.PostSettings.Tone,
				ReferenceId: utils.StringPtrOrNil(h.PostSettings.ReferenceID),
			},
		})
	}

	return &pbcore.PostMetadata{
		Settings: &pbcore.PostSettings{
			Topic:       meta.Settings.Topic,
			Context:     meta.Settings.Context,
			Goal:        meta.Settings.Goal,
			Tone:        meta.Settings.Tone,
			ReferenceId: utils.StringPtrOrNil(meta.Settings.ReferenceID),
		},
		History: history,
	}
}

// Main converter: models.Post → pbcore.Post
func FromModel(post *models.Post) *pbcore.Post {
	return &pbcore.Post{
		Id:          post.ID,
		ProjectId:   post.ProjectID,
		Source:      post.SourceID,
		Topic:       post.Title,
		Description: post.Description,
		Status:      string(post.Status),
		Reason:      post.Reason,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		ScheduledAt: utils.TimestamppbOrNil(post.ScheduleAt),
		Metadata:    convertMetadata(post.Metadata),
	}
}

func (i *IntegrationType) FromModel(model models.IntegrationType) {
	value := "INTEGRATION_TYPE_" + strings.ToUpper(model.String())
	enum, found := IntegrationType_value[value]
	if !found {
		panic(fmt.Errorf("unknown integration type model %q", model))
	}

	*i = IntegrationType(enum)
}

func (c IntegrationType) ToModel() models.IntegrationType {
	value := strings.TrimPrefix(strings.ToUpper(c.String()), "INTEGRATION_TYPE_")
	model := models.IntegrationType(value)
	if !model.IsValid() {
		panic(fmt.Errorf("unknown integration type pb %q", value))
	}

	return model
}

func (i *IntegrationState) FromModel(model models.IntegrationState) {
	value := "INTEGRATION_STATE_" + strings.ToUpper(model.String())
	enum, found := IntegrationState_value[value]
	if !found {
		panic(fmt.Errorf("unknown integration state model %q", model))
	}

	*i = IntegrationState(enum)
}

func (i *Integration) FromModel(model *models.Integration) *Integration {
	i.Id = model.ID
	i.OrganizationId = model.OrganizationID
	i.Type.FromModel(model.Type)
	i.Status.FromModel(model.State)
	return i
}
