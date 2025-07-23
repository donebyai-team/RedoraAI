package interactions

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"strings"
)

type DMParams struct {
	ID         string
	Cookie     string // json array
	To         string
	ToUsername string
	Message    string
}

const disabledReasonAccNotEstablished = "Your Reddit account hasn't met Reddit's requirements for sending direct messages, which typically include email verification and a history of positive contributions."

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

	err = r.redditOauthClient.WithRotatingAccounts(ctx, interaction.Organization.ID, models.IntegrationTypeREDDITDMLOGIN, func(integration *models.Integration) error {
		config := integration.GetRedditDMLoginConfig()
		interaction.From = config.Username

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

		user, err := reddit.NewClientWithOutConfig(r.logger).GetUser(ctx, interaction.To)
		if err != nil {
			interaction.Status = models.LeadInteractionStatusFAILED
			interaction.Reason = fmt.Sprintf("Reason: %v", err)
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

		updatedCookies, err := r.browserLessClient.SendDM(ctx, DMParams{
			ID:         interaction.ID,
			Cookie:     config.Cookies,
			To:         fmt.Sprintf("t2_%s", user.ID),
			ToUsername: user.Name,
			Message:    utils.FormatDM(redditLead.LeadMetadata.SuggestedDM),
		})
		if err != nil {
			interaction.Reason = fmt.Sprintf("Reason: %v", err)
			interaction.Status = models.LeadInteractionStatusFAILED
			if strings.Contains(err.Error(), "account isn't established") {
				interaction.Reason = disabledReasonAccNotEstablished
			}

			return err
		}

		config.Cookies = string(updatedCookies)
		integration = models.SetIntegrationType(integration, models.IntegrationTypeREDDITDMLOGIN, config)
		_, err = r.db.UpsertIntegration(ctx, integration)
		if err != nil {
			interaction.Reason = fmt.Sprintf("failed to update integration: %v", err)
			interaction.Status = models.LeadInteractionStatusFAILED
			return err
		}

		interaction.Status = models.LeadInteractionStatusSENT
		interaction.Reason = ""
		redditLead.LeadMetadata.AutomatedDMSent = true
		redditLead.Status = models.LeadStatusAIRESPONDED

		r.logger.Info("successfully sent reddit DM",
			zap.String("interaction_id", interaction.ID),
			zap.String("from", interaction.From))

		return nil
	})

	if err != nil {
		interaction.Status = models.LeadInteractionStatusFAILED
		// if the reason is not set then set it to the error message
		if interaction.Reason == "" {
			interaction.Reason = err.Error()
		}

		if errors.Is(err, reddit.AllAccountBanned) {
			r.disableAutomation(ctx, interaction, reddit.AllAccountBanned.Error())
		} else if errors.Is(err, reddit.AllAccountNotEstablished) {
			r.disableAutomation(ctx, interaction, reddit.AllAccountNotEstablished.Error())
		}
	}

	return err
}

type loginCallback func() error

func (r redditInteractions) Authenticate(ctx context.Context, orgID string, cookieJSON string) (string, loginCallback, error) {
	// Handle direct cookie login
	if cookieJSON != "" {
		updatedLoginConfig, err := r.browserLessClient.ValidateCookies(ctx, cookieJSON)
		if err != nil {
			return "", nil, err
		}

		if err := r.finalizeLogin(ctx, orgID, updatedLoginConfig); err != nil {
			return "", nil, err
		}

		r.logger.Info("successfully logged in to reddit", zap.String("org_id", orgID))
		return "", nil, nil
	}

	// Handle login via browser automation
	cdp, err := r.browserLessClient.StartLogin(ctx)
	if err != nil {
		return "", nil, err
	}

	return cdp.LiveURL, func() error {
		updatedLoginConfig, err := r.browserLessClient.WaitAndGetCookies(ctx, cdp.BrowserWSEndpoint)
		if err != nil {
			return err
		}

		if err := r.finalizeLogin(ctx, orgID, updatedLoginConfig); err != nil {
			return err
		}

		r.logger.Info("successfully logged in to reddit", zap.String("org_id", orgID))
		return nil
	}, nil
}

func (r redditInteractions) finalizeLogin(ctx context.Context, orgID string, updatedLoginConfig *models.RedditDMLoginConfig) error {
	if updatedLoginConfig.Username == "" {
		return fmt.Errorf("unable to find username in cookies")
	}

	// Validate Reddit user is still active
	if _, err := reddit.NewClientWithOutConfig(r.logger).GetUser(ctx, updatedLoginConfig.Username); err != nil {
		return err
	}

	integration := &models.Integration{
		OrganizationID: orgID,
		State:          models.IntegrationStateACTIVE,
		Type:           models.IntegrationTypeREDDITDMLOGIN,
	}

	integration = models.SetIntegrationType(integration, models.IntegrationTypeREDDITDMLOGIN, updatedLoginConfig)

	if _, err := r.db.UpsertIntegration(ctx, integration); err != nil {
		r.logger.Warn("failed to update integration", zap.Error(err), zap.String("org_id", orgID))
		return err
	}

	return nil
}
