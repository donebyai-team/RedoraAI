package pbcore

import (
	"fmt"
	"github.com/shank318/doota/models"
	"strings"
)

func (c SubscriptionPlanID) ToModel() models.SubscriptionPlanType {
	value := strings.TrimPrefix(strings.ToUpper(c.String()), "SUBSCRIPTION_PLAN_")
	model := models.SubscriptionPlanType(value)
	if !model.IsValid() {
		panic(fmt.Errorf("unknown subscription pla type pb %q", value))
	}

	return model
}
