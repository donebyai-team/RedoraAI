package api

type CreateCallRequest struct {
	FromPhone               string            `json:"from_phone"`
	ToPhone                 string            `json:"to_phone"`
	AgentID                 string            `json:"agent_id"`
	IncludeMetadataInPrompt string            `json:"include_metadata_in_prompt"`
	Metadata                map[string]string `json:"metadata"`
}

type CreateCallResponse struct {
	CallID    string `json:"call_id"`
	SessionID string `json:"session_id"`
}
