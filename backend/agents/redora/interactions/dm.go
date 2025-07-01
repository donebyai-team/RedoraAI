package interactions

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"strings"
)

type DMParams struct {
	ID       string
	Username string
	Password string
	Cookie   string // json array
	To       string
	Message  string
}

func (r redditInteractions) SendDM(ctx context.Context, interaction *models.LeadInteraction) error {
	if interaction.Type != models.LeadInteractionTypeDM {
		return fmt.Errorf("interaction type is not DM")
	}
	r.logger.Info("sending reddit DM",
		zap.String("interaction_id", interaction.ID),
		zap.String("from", interaction.From))

	project, err := r.db.GetProject(ctx, interaction.ProjectID)
	if err != nil {
		return err
	}

	redditLead, err := r.db.GetLeadByID(ctx, interaction.ProjectID, interaction.LeadID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(utils.FormatDM(redditLead.LeadMetadata.SuggestedDM)) == "" {
		return fmt.Errorf("no DM message found")
	}

	defer func() {
		// Always update interaction at the end
		updateErr := r.db.UpdateLeadInteraction(ctx, interaction)
		if updateErr != nil && err == nil {
			err = fmt.Errorf("failed to update interaction: %w", updateErr)
		}

		redditLead.LeadMetadata.DMScheduledAt = nil
		updateError := r.db.UpdateLeadStatus(ctx, redditLead)
		if updateError != nil {
			r.logger.Warn("failed to update lead status for automated DM", zap.Error(err), zap.String("lead_id", redditLead.ID))
		}
	}()

	if !interaction.Organization.FeatureFlags.IsSubscriptionActive() {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "subscription has expired or not active"
		return nil
	}

	if !project.IsActive {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "project is not active"
		return nil
	}

	if redditLead.Status == models.LeadStatusNOTRELEVANT {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "Skipped, as user marked it as not relevant"
		return nil
	}

	if redditLead.Status == models.LeadStatusCOMPLETED {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "Skipped, as user has marked it responded manually"
		return nil
	}

	if strings.TrimSpace(utils.FormatDM(redditLead.LeadMetadata.SuggestedDM)) == "" {
		err := fmt.Errorf("no DM message found")
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = err.Error()
		return err
	}

	// case: if auto DM disabled
	if !interaction.Organization.FeatureFlags.EnableAutoDM {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "auto DM is disabled for this organization"
		return nil
	}

	// Check the interaction should not exist
	exists, err := r.db.IsInteractionExists(ctx, interaction)
	if err != nil {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = fmt.Sprintf("failed to check if interaction exists: %s", err.Error())
		return err
	}

	if exists {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "DM already exists"
		return nil
	}

	integration, err := r.db.GetIntegrationByOrgAndType(ctx, interaction.Organization.ID, models.IntegrationTypeREDDITDMLOGIN)
	if err != nil && errors.Is(err, datastore.NotFound) {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "integration not found or inactive"
		return err
	}

	if integration.State != models.IntegrationStateACTIVE {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "dm integration not found or inactive"
		return fmt.Errorf(interaction.Reason)
	}

	redditClient, err := r.redditOauthClient.GetOrCreate(ctx, interaction.Organization.ID, false)
	if err != nil {
		if errors.Is(err, datastore.IntegrationNotFoundOrActive) {
			interaction.Status = models.LeadInteractionStatusFAILED
			interaction.Reason = "oauth integration not found or inactive"
		} else {
			interaction.Status = models.LeadInteractionStatusFAILED
			interaction.Reason = err.Error()
		}
		return err
	}

	user, err := redditClient.GetUser(ctx, interaction.To)
	if err != nil {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = fmt.Sprintf("failed to get user: %v", err)
		return err
	}

	if user == nil || user.ID == "" {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "user does not exist or suspended"
		return nil
	}

	err = r.db.SetLeadInteractionStatusProcessing(ctx, interaction.ID)
	if err != nil {
		return err
	}

	loginConfig := integration.GetRedditDMLoginConfig()

	updatedCookies, err := r.browserLessClient.SendDM(ctx, DMParams{
		ID:       interaction.ID,
		Cookie:   loginConfig.Cookies,
		Username: loginConfig.Username,
		Password: loginConfig.Password,
		To:       fmt.Sprintf("t2_%s", user.ID),
		Message:  utils.FormatDM(redditLead.LeadMetadata.SuggestedDM),
	})
	if err != nil {
		interaction.Reason = fmt.Sprintf("failed to send DM: %v", err)
		interaction.Status = models.LeadInteractionStatusFAILED
		return err
	}

	loginConfig.Cookies = string(updatedCookies)
	_, err = r.db.UpsertIntegration(ctx, integration)
	if err != nil {
		interaction.Reason = fmt.Sprintf("failed to update integration: %v", err)
		interaction.Status = models.LeadInteractionStatusFAILED
		return err
	}

	interaction.Status = models.LeadInteractionStatusSENT
	interaction.Reason = ""

	if loginConfig.Username != "" {
		interaction.From = loginConfig.Username
	}
	redditLead.LeadMetadata.AutomatedDMSent = true
	redditLead.Status = models.LeadStatusAIRESPONDED

	r.logger.Info("successfully sent reddit DM",
		zap.String("interaction_id", interaction.ID),
		zap.String("from", interaction.From))

	return nil
}

type loginCallback func() error

func (r redditInteractions) Authenticate(ctx context.Context, orgID string) (string, loginCallback, error) {
	cdp, err := r.browserLessClient.StartLogin(ctx)
	if err != nil {
		return "", nil, err
	}

	return cdp.LiveURL, func() error {
		updatedLoginConfig, err := r.browserLessClient.WaitAndGetCookies(ctx, cdp.BrowserWSEndpoint)
		if err != nil {
			return err
		}

		integration := &models.Integration{
			OrganizationID: orgID,
			State:          models.IntegrationStateACTIVE,
			Type:           models.IntegrationTypeREDDITDMLOGIN,
		}

		integration = models.SetIntegrationType(integration, models.IntegrationTypeREDDITDMLOGIN, updatedLoginConfig)
		_, err = r.db.UpsertIntegration(ctx, integration)
		if err != nil {
			r.logger.Warn("failed to update integration", zap.Error(err), zap.String("org_id", orgID))
			return err
		}
		r.logger.Info("successfully logged in to reddit", zap.String("org_id", orgID))
		return nil
	}, nil
}
