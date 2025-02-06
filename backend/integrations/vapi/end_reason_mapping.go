package vapi

import (
	api "github.com/VapiAI/server-sdk-go"
	"github.com/shank318/doota/models"
)

var endReasonMapping = map[models.CallEndedReason][]api.CallEndedReason{
	models.CallEndedReasonASSISTANTENDED: {
		api.CallEndedReasonAssistantEndedCall,
		api.CallEndedReasonAssistantSaidEndCallPhrase,
		api.CallEndedReasonSilenceTimedOut,
	},
	models.CallEndedReasonASSISTANTFORWARDED: {
		api.CallEndedReasonAssistantForwardedCall,
	},
	models.CallEndedReasonCUSTOMERBUSY: {
		api.CallEndedReasonCustomerBusy,
		api.CallEndedReasonCustomerDidNotAnswer,
		api.CallEndedReasonCustomerDidNotGiveMicrophonePermission,
		api.CallEndedReasonVoicemail,
	},
	models.CallEndedReasonCUSTOMERENDED: {
		api.CallEndedReasonCustomerEndedCall,
		api.CallEndedReasonManuallyCanceled,
	},
	models.CallEndedReasonMAXCALLDURATIONREACHED: {
		api.CallEndedReasonExceededMaxDuration,
	},
}

var callStatusMapping = map[models.CallStatus][]api.CallStatus{
	models.CallStatusQUEUED: {
		api.CallStatusQueued,
	},
	models.CallStatusINPROGRESS: {
		api.CallStatusInProgress,
		api.CallStatusRinging,
	},
	models.CallStatusENDED: {
		api.CallStatusEnded,
	},
	models.CallStatusFORWARDED: {
		api.CallStatusForwarding,
	},
}
