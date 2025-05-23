package interactions

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/errorx"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type DMParams struct {
	ID       string
	Username string
	Password string
	To       string
	Message  string
}

func (r redditInteractions) SendDM(ctx context.Context, interaction *models.LeadInteraction) error {
	if interaction.Type != models.LeadInteractionTypeDM {
		return fmt.Errorf("interaction type is not DM")
	}
	r.logger.Info("sending reddit DM", zap.String("from", interaction.From))

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

	integration, err := r.db.GetIntegrationByOrgAndType(ctx, interaction.Organization.ID, models.IntegrationTypeREDDITDMLOGIN)
	if err != nil && errors.Is(err, datastore.NotFound) {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "integration not found or inactive"
		return err
	}

	redditClient, err := r.redditOauthClient.GetOrCreate(ctx, interaction.Organization.ID, true)
	if err != nil {
		if errors.Is(err, datastore.IntegrationNotFoundOrActive) {
			interaction.Status = models.LeadInteractionStatusFAILED
			interaction.Reason = "integration not found or inactive"
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

	if err = r.browserLessClient.SendDM(DMParams{
		ID:       interaction.ID,
		Username: loginConfig.Username,
		Password: loginConfig.Password,
		To:       fmt.Sprintf("t2_%s", user.ID),
		Message:  utils.FormatDM(redditLead.LeadMetadata.SuggestedDM),
	}); err != nil {
		interaction.Reason = fmt.Sprintf("failed to send DM: %v", err)
		interaction.Status = models.LeadInteractionStatusFAILED
		return err
	}

	interaction.Status = models.LeadInteractionStatusSENT
	interaction.Reason = ""
	redditLead.LeadMetadata.AutomatedDMSent = true
	redditLead.Status = models.LeadStatusCOMPLETED

	return nil
}

func (r redditInteractions) CheckIfLogin(ctx context.Context, orgID string) error {
	integration, err := r.db.GetIntegrationByOrgAndType(ctx, orgID, models.IntegrationTypeREDDITDMLOGIN)
	if err != nil {
		return err
	}

	loginConfig := integration.GetRedditDMLoginConfig()
	err = r.browserLessClient.CheckIfLogin(DMParams{
		ID:       integration.ID,
		Username: loginConfig.Username,
		Password: loginConfig.Password,
	})
	var loginErr *errorx.LoginError
	if err != nil && errors.As(err, &loginErr) {
		r.logger.Warn("failed to login to reddit", zap.Error(err), zap.String("org_id", orgID))
		return status.Error(codes.InvalidArgument, loginErr.Reason)
	} else if err != nil {
		r.logger.Warn("failed to login to reddit", zap.Error(err), zap.String("org_id", orgID))
		return status.Error(codes.Internal, "unable login to reddit")
	}

	return nil
}
