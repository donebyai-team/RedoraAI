package portal

import (
	"github.com/gorilla/mux"
	"github.com/shank318/doota/agents"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dhttp"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

func (p *Portal) getWebhookHandler(agent agents.AIAgent) agents.WebhookHandler {
	if agent == agents.AIAgentVANA {
		return p.vanaWebhookHandler
	}
	return nil
}

func (p *Portal) UpdateCallStatusHandler(agent agents.AIAgent) http.HandlerFunc {
	webhookHandlerToUse := p.getWebhookHandler(agent)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.Logger(ctx, p.logger)

		logger.Debug("received update call status request", zap.String("path", r.URL.Path))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			dhttp.WriteError(r.Context(), w, derr.UnexpectedError(r.Context(), err))
			return
		}
		defer r.Body.Close() // Close the body after reading

		vars := mux.Vars(r)
		conversationID, ok := vars["id"]
		if !ok {
			dhttp.WriteError(r.Context(), w, derr.RequestValidationError(ctx, url.Values{"path": []string{"expected conversation id"}}))
			return
		}
		logger.Info("handling update call status request", zap.String("conversation_id", conversationID))
		err = webhookHandlerToUse.UpdateCallStatus(r.Context(), conversationID, body)
		if err != nil {
			dhttp.WriteError(r.Context(), w, derr.UnexpectedError(r.Context(), err))
		}
	}
}

func (p *Portal) EndConversationHandler(agent agents.AIAgent) http.HandlerFunc {
	webhookHandlerToUse := p.getWebhookHandler(agent)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.Logger(ctx, p.logger)

		logger.Debug("received end conversation request", zap.String("path", r.URL.Path))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			dhttp.WriteError(r.Context(), w, derr.UnexpectedError(r.Context(), err))
			return
		}
		defer r.Body.Close() // Close the body after reading

		vars := mux.Vars(r)
		conversationID, ok := vars["id"]
		if !ok {
			dhttp.WriteError(r.Context(), w, derr.RequestValidationError(ctx, url.Values{"path": []string{"expected conversation id"}}))
			return
		}
		logger.Info("handling end conversation request", zap.String("conversation_id", conversationID))
		err = webhookHandlerToUse.EndConversation(r.Context(), conversationID, body)
		if err != nil {
			dhttp.WriteError(r.Context(), w, derr.UnexpectedError(r.Context(), err))
		}
	}
}
