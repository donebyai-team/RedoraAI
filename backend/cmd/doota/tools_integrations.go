package main

import (
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/app"
	"github.com/shank318/doota/models"
	"github.com/spf13/cobra"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	"github.com/tidwall/gjson"
)

var toolsIntegrationsGroup = Group(
	"integrations",
	"Commands related to a integrations store",
	toolsDat,
	toolsGet,
	toolsSlackWebhook,
	toolsRedditLoginConfig,
)

var toolsDat = Group(
	"vapi",
	"Will insert the vapi config in integrations db for a tenant",
	toolsDatCreate,
)

var toolsSlackWebhook = Group(
	"slack_webhook",
	"Will insert the Slack webhook config in integrations db for a tenant",
	toolsSlackWebhookCreate,
)

var toolsRedditLoginConfig = Group(
	"reddit_login_config",
	"Will insert the reddit login config in integrations db for a tenant",
	toolsRedditLoginCreate,
)

var toolsGet = Command(
	toolsGetRunE,
	"get",
	"retrieve an integration",
)

var toolsDatCreate = Command(
	toolsCreateIntegration,
	"create <tenant-id> <config>",
	"Will insert the dat config in integrations db",
)

var toolsSlackWebhookCreate = Command(
	toolsCreateIntegrationSlackWebhook,
	"create <tenant-id> <config>",
	"Will insert the slack webhook config in integrations db")

var toolsRedditLoginCreate = Command(
	toolsCreateIntegrationRedditLogin,
	"create <tenant-id> <config>",
	"Will insert the reddit login config in integrations db")

func toolsGetRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	integrationID := args[0]

	db, err := app.SetupDataStore(ctx, sflags.MustGetString(cmd, "pg-dsn"), zlog, tracer)
	if err != nil {
		return fmt.Errorf("failed to setup datastore: %w", err)
	}

	fmt.Println("Getting integration: ", integrationID)

	integration, err := db.GetIntegrationById(ctx, integrationID)
	if err != nil {
		return err
	}

	cnt, err := json.MarshalIndent(integration, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(cnt))

	switch integration.Type {
	case models.IntegrationTypeVOICEVAPI:
		fmt.Println("Vapi Config: ")
		config := integration.GetVAPIConfig()
		fmt.Println("   APIKEY: ", config.APIKey)
		fmt.Println("   APIHost: ", config.HostName)
		return nil
	case models.IntegrationTypeREDDIT:
		fmt.Println("Reddit Config: ")
		config := integration.GetRedditConfig()
		fmt.Println("   Token: ", config.AccessToken)
		fmt.Println("   UserName: ", config.Name)
		return nil
	}
	return nil
}

func toolsCreateIntegration(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, err := app.SetupDataStore(ctx, sflags.MustGetString(cmd, "pg-dsn"), zlog, tracer)
	if err != nil {
		return fmt.Errorf("failed to setup datastore: %w", err)
	}

	integration := &models.Integration{
		OrganizationID: args[0],
		State:          models.IntegrationStateACTIVE,
		Type:           models.IntegrationTypeVOICEVAPI,
	}

	out := &models.VAPIConfig{}
	if err := json.Unmarshal([]byte(args[1]), out); err != nil {
		return fmt.Errorf("unable to unmarshal dat config: %w", err)
	}
	// we need to chery pick the password since it is not exposed via json interface
	out.APIKey = gjson.Get(args[1], "api_key").String()

	integration = models.SetIntegrationType(integration, models.IntegrationTypeVOICEVAPI, out)
	upsertIntegration, err := db.UpsertIntegration(ctx, integration)
	if err != nil {
		return err
	}

	fmt.Println("Integration Created: ", upsertIntegration.ID)
	data, _ := json.MarshalIndent(upsertIntegration, "", "  ")
	fmt.Println(string(data))

	return nil
}

func toolsCreateIntegrationRedditLogin(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, err := app.SetupDataStore(ctx, sflags.MustGetString(cmd, "pg-dsn"), zlog, tracer)
	if err != nil {
		return fmt.Errorf("failed to setup datastore: %w", err)
	}

	integration := &models.Integration{
		OrganizationID: args[0],
		State:          models.IntegrationStateACTIVE,
		Type:           models.IntegrationTypeREDDITDMLOGIN,
	}

	out := &models.RedditDMLoginConfig{}
	if err := json.Unmarshal([]byte(args[1]), out); err != nil {
		return fmt.Errorf("unable to unmarshal dat config: %w", err)
	}
	// we need to chery pick the password since it is not exposed via json interface
	out.Password = gjson.Get(args[1], "password").String()

	integration = models.SetIntegrationType(integration, models.IntegrationTypeREDDITDMLOGIN, out)
	upsertIntegration, err := db.UpsertIntegration(ctx, integration)
	if err != nil {
		return err
	}

	fmt.Println("Integration Created: ", upsertIntegration.ID)
	data, _ := json.MarshalIndent(upsertIntegration, "", "  ")
	fmt.Println(string(data))

	return nil
}

func toolsCreateIntegrationSlackWebhook(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, err := app.SetupDataStore(ctx, sflags.MustGetString(cmd, "pg-dsn"), zlog, tracer)
	if err != nil {
		return fmt.Errorf("failed to setup datastore: %w", err)
	}

	integration := &models.Integration{
		OrganizationID: args[0],
		State:          models.IntegrationStateACTIVE,
		Type:           models.IntegrationTypeSLACKWEBHOOK,
	}

	out := &models.SlackWebhookConfig{}
	if err := json.Unmarshal([]byte(args[1]), out); err != nil {
		return fmt.Errorf("unable to unmarshal dat config: %w", err)
	}
	// we need to chery pick the password since it is not exposed via json interface
	out.Webhook = gjson.Get(args[1], "webhook").String()

	integration = models.SetIntegrationType(integration, models.IntegrationTypeSLACKWEBHOOK, out)
	upsertIntegration, err := db.UpsertIntegration(ctx, integration)
	if err != nil {
		return err
	}

	fmt.Println("Integration Created: ", upsertIntegration.ID)
	data, _ := json.MarshalIndent(upsertIntegration, "", "  ")
	fmt.Println(string(data))

	return nil
}
