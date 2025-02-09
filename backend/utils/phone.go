package utils

import "fmt"
import "github.com/nyaruka/phonenumbers"

// ConvertToE164 formats a phone number into E.164 format
func ConvertToE164(phoneNumber, region string) (string, error) {
	parsedNumber, err := phonenumbers.Parse(phoneNumber, region)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %v", err)
	}

	// Format the number in E.164 format
	return phonenumbers.Format(parsedNumber, phonenumbers.E164), nil
}
