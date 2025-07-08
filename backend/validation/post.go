package validation

import (
	"errors"
	"fmt"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type SchedulePostInput struct {
	ID         string    `validate:"required,uuid4"`
	ScheduleAt time.Time `validate:"required"`
}

func ValidateSchedulePost(req *pbcore.SchedulePostRequest, post *models.Post, expectedProjectID string) error {
	input := SchedulePostInput{
		ID:         req.GetId(),
		ScheduleAt: req.GetScheduleAt().AsTime(),
	}

	// Step 1: Struct tag-based validation
	if err := validate.Struct(input); err != nil {
		return formatValidationError(err)
	}

	//Business logic validation
	if post.ProjectID != expectedProjectID {
		return errors.New("you do not have access to this post")
	}

	if post.ScheduleAt != nil {
		return errors.New("post is already scheduled")
	}

	if input.ScheduleAt.Before(time.Now()) {
		return errors.New("schedule time cannot be in the past")
	}

	return nil
}

func formatValidationError(err error) error {
	if ve, ok := err.(validator.ValidationErrors); ok {
		var combinedErr string
		for _, e := range ve {
			combinedErr += fmt.Sprintf("Field '%s' failed validation on '%s'\n", e.Field(), e.Tag())
		}
		return errors.New(combinedErr)
	}
	return err
}
