package agents

import "github.com/shank318/doota/models"

func IsCallRunning(status models.CallStatus) bool {
	return status == models.CallStatusQUEUED ||
		status != models.CallStatusINPROGRESS
}
