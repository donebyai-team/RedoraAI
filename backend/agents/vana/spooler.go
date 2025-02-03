package vana

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Spooler struct {
	*shutter.Shutter
	dbPollingInterval time.Duration

	db                 datastore.Repository
	aiClient           ai.Client
	queue              chan *models.AugmentedCustomerCase
	queued             *agents.QueuedMap[string, bool]
	integrationFactory integrations.Factory
	appIsReady         func() bool

	logger *zap.Logger
}

func New(
	db datastore.Repository,
	aiClient ai.Client,
	integrationFactory integrations.Factory,
	bufferSize int,
	dbPollingInterval time.Duration,
	isShuttingDown func() bool,
	logger *zap.Logger,
) *Spooler {
	return &Spooler{
		Shutter:            shutter.New(),
		db:                 db,
		aiClient:           aiClient,
		integrationFactory: integrationFactory,
		dbPollingInterval:  dbPollingInterval,
		appIsReady:         isShuttingDown,
		queue:              make(chan *models.AugmentedCustomerCase, bufferSize),
		queued:             agents.NewQueuedMap[string, bool](bufferSize),
		logger:             logger,
	}
}

func (s *Spooler) Run(ctx context.Context) error {
	go s.runLoop(ctx)
	go s.pollCustomerCases(ctx)
	return nil
}

func (s *Spooler) runLoop(ctx context.Context) {
	s.logger.Info("running spooler loop")
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("ctx done, run loop ending")
			return
		case <-s.Terminating():
			s.logger.Info("spooler terminating, run loop ending")
			return
		case customerCase := <-s.queue:
			// Remove the case from the queued map, we are processing it
			s.queued.Delete(customerCase.CustomerCase.ID)

			// FIXME: We need to deal with errors differently here. We need to separated
			// internal spooler error that are irecoverable from the ones that are
			// coming from the investigator or are recoverable.
			//
			// Indeed, we don't want to stop the spooler if the investigator is broken or
			// something.
			if err := s.processCustomerCase(ctx, customerCase); err != nil {
				s.Shutdown(fmt.Errorf("process investigation: %w", err))
				return
			}
		}
	}
}

const fromPhone = ""

func (s *Spooler) processCustomerCase(ctx context.Context, customerCase *models.AugmentedCustomerCase) error {
	logger := s.logger.With(
		zap.String("customer_case_id", customerCase.CustomerCase.ID),
		zap.String("customer_id", customerCase.Customer.ID),
		zap.String("organization_id", customerCase.CustomerCase.OrgID),
		zap.String("creator", "vana"),
	)
	logger.Debug("processing customer cases", zap.Int("queue_size", len(s.queue)))

	// Create Conversation
	voiceProvider, err := s.integrationFactory.NewVoiceClient(ctx, customerCase.CustomerCase.OrgID)
	if err != nil {
		return err
	}

	promptConfig, err := s.db.GetPromptTypeByName(ctx, customerCase.CustomerCase.OrgID, customerCase.CustomerCase.PromptType)
	if err != nil {
		return fmt.Errorf("failed to get prompt type for %s: %w", customerCase.CustomerCase.OrgID, err)
	}

	prompt := ai.Prompt{}
	if err := json.Unmarshal(promptConfig.Config, &prompt); err != nil {
		return fmt.Errorf("unmarshal extractor config into prompt: %w", err)
	}

	conversation, err := s.db.CreateConversation(ctx, &models.Conversation{
		CustomerCaseID: customerCase.CustomerCase.ID,
		FromPhone:      fromPhone,
		Provider:       voiceProvider.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to create conversation %q: %w", customerCase.CustomerCase.ID, err)
	}

	debugTemplateName := fmt.Sprintf("voice.%s", strings.ToLower(promptConfig.Name))

	vars := prompt.Model.GetVars(customerCase)
	chatMessages, err := s.aiClient.ExtractMessages(ctx, debugTemplateName, prompt, vars, conversation.ID, logger)
	if err != nil {
		return fmt.Errorf("failed to extract messages from prompt: %w", err)
	}

	callResponse, err := voiceProvider.CreateCall(ctx, models.CallRequest{
		ConversationID: conversation.ID,
		FromPhone:      conversation.FromPhone,
		ToPhone:        customerCase.Customer.Phone,
		ChatMessages:   chatMessages,
		GPTModel:       prompt.Model.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to create call response for %s: %w", customerCase.CustomerCase.ID, err)
	}

	conversation.ExternalID = callResponse.CallID
	conversation.CallStatus = callResponse.Status

	err = s.db.UpdateConversation(ctx, conversation)
	if err != nil {
		return fmt.Errorf("failed to update conversation for %s: %w", customerCase.CustomerCase.ID, err)
	}

	return nil
}

func (s *Spooler) pollCustomerCases(ctx context.Context) {
	// 0 so the first time we poll, we do it right away
	interval := 0 * time.Second
	location, _ := time.LoadLocation("Asia/Kolkata") // IST timezone

	for {
		select {
		case <-time.After(interval):
			now := time.Now().In(location)
			if now.Hour() >= 10 && now.Hour() < 18 { // Run only between 10 AM and 6 PM IST
				if err := s.loadCustomerSessions(ctx); err != nil {
					s.Shutdown(fmt.Errorf("fail to load customer sessions from db: %w", err))
				}
			}
		case <-ctx.Done():
		}

		// If we have 0 it means we just started, move to the real interval now
		if interval == 0 {
			interval = s.dbPollingInterval
		}
	}
}

func (s *Spooler) loadCustomerSessions(ctx context.Context) error {
	t0 := time.Now()
	// Case IN (CREATED, PENDING)
	// LastCallStatus = NULL OR IN (ENDED, AI_ENDED)
	// NextScheduledAt = NULL or <= current time
	cases, err := s.db.GetCustomerCases(ctx, datastore.CustomerCaseFilter{
		CaseStatus:      []models.CustomerCaseStatus{models.CustomerCaseStatusCREATED, models.CustomerCaseStatusPENDING},
		LastCallStatus:  []models.CallStatus{models.CallStatusENDED, models.CallStatusAIENDED},
		NextScheduledAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("processing customer cases: %w", err)
	}

	casesToProcess := 0
	for _, customerCase := range cases {
		if !s.shouldNotProcessCustomerCase(customerCase) {
			casesToProcess++
			s.pushCustomerSession(customerCase)
		}
	}
	s.logger.Info("found cases to process from db", zap.Int("count", casesToProcess), zap.Duration("elapsed", time.Since(t0)))
	return nil
}

func (s *Spooler) shouldNotProcessCustomerCase(customerCase *models.AugmentedCustomerCase) bool {
	return (customerCase.CustomerCase.Status == models.CustomerCaseStatusCLOSED ||
		customerCase.CustomerCase.Status == models.CustomerCaseStatusPAID ||
		customerCase.CustomerCase.Status == models.CustomerCaseStatusPARTIALLYPAID) &&
		(customerCase.CustomerCase.LastCallStatus == models.CallStatusQUEUED ||
			customerCase.CustomerCase.LastCallStatus == models.CallStatusINPROGRESS)
}

func (s *Spooler) pushCustomerSession(customerCase *models.AugmentedCustomerCase) {
	if s.queued.Has(customerCase.CustomerCase.ID) {
		return
	}

	// TODO should we check size vs buffer?
	s.queue <- customerCase
	s.queued.Set(customerCase.CustomerCase.ID, true)
}
