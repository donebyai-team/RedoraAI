package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shank318/doota/prompttypes"
	"os"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/shank318/doota/app"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/spf13/cobra"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	"github.com/test-go/testify/assert"
)

var toolsPTSGroup = Group(
	"pts",
	"Commands related to a prompt type store",
	toolsPTSSyncCmd,
)

var toolsPTSSyncCmd = Command(
	toolsMTSSyncRunE,
	"sync",
	"Will sync the message type store with a database, update and creating  message types and extractor, where appropriate",
)

func getStore(cmd *cobra.Command) (*prompttypes.Store, error) {
	storeUrl := sflags.MustGetString(cmd, "message-type-store-url")
	if storeUrl == "" {
		return nil, fmt.Errorf("message-type-store-url is required")
	}
	reader, err := prompttypes.NewReader(storeUrl, zlog)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}
	store, err := reader.Reader(cmd.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to read store: %w", err)
	}

	return store, nil
}

func toolsMTSSyncRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	mtStore, err := getStore(cmd)
	if err != nil {
		return err
	}

	zlog.Info("generating prompt type")

	db, err := app.SetupDataStore(ctx, sflags.MustGetString(cmd, "pg-dsn"), zlog, tracer)
	if err != nil {
		return fmt.Errorf("failed to setup datastore: %w", err)
	}

	fmt.Println("******************************************")
	fmt.Println("************ Prompt type setup **********")
	fmt.Println("******************************************")
	fmt.Println("")

	run := false
	dds, err := promptConfirm("Do you want to run in LIVE mode and and affect the DB", &promptOptions{
		PromptTemplates: &promptui.PromptTemplates{
			Success: `{{ "Run in live mode:" | faint }} `,
		},
	})
	if err != nil {
		return err
	}
	if dds {
		run = true
	}

	orgs, err := db.GetOrganizations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organizations: %w", err)
	}

	syncer := &PromptTypeSyncer{
		db:      db,
		execute: run,
		store:   mtStore,
		orgs:    orgs,
		count:   &count{},
	}
	if err := syncer.run(ctx); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Summary")
	fmt.Println("Total:", syncer.count.Total)
	fmt.Println("  > no op operations:", syncer.count.Noop)
	fmt.Println("  > out of sync operations:", syncer.count.OutOfSync)
	fmt.Println("  > create operations:", syncer.count.Create)

	return nil
}

type PromptTypeSyncer struct {
	execute bool
	db      datastore.Repository
	org     *models.Organization
	store   *prompttypes.Store
	orgs    []*models.Organization
	count   *count
}

type count struct {
	Total     int `json:"total"`
	OutOfSync int `json:"out_of_sync"`
	Create    int `json:"create"`
	Noop      int `json:"noop"`
}

func (m *PromptTypeSyncer) run(ctx context.Context) error {
	for _, mt := range m.store.PromptTypes() {
		if err := m.sync(ctx, mt.Name); err != nil {
			return err
		}
	}
	return nil
}

func (m *PromptTypeSyncer) sync(ctx context.Context, messageName string) error {

	messageType := m.store.MustGetPromptType(messageName)

	fmt.Println("")
	fmt.Println("Syncing message type: ", messageType.Name)

	m.count.Total++
	if err := m.syncMessageType(ctx, messageType); err != nil {
		return fmt.Errorf("failed to sync message type %s: %w", messageType.Name, err)
	}

	return nil
}

func (m *PromptTypeSyncer) syncMessageType(ctx context.Context, promptType *models.PromptType) error {
	existingPromptType, err := m.db.GetPromptTypeByName(ctx, promptType.Name)
	if err != nil && err != datastore.NotFound {
		return fmt.Errorf("failed to get prompt source: %w", err)
	}
	if err == datastore.NotFound {
		m.count.Create++
		fmt.Printf("  > ⚠️ Prompt Type Not Found creating it\n")
		if !m.execute {
			return nil
		}

		org, err := m.selectOrg()
		if err != nil {
			return fmt.Errorf("failed to select organization: %w", err)
		}

		if org != nil {
			promptType.OrganizationId = &org.ID
		}

		promptType, err = m.db.CreatePromptType(ctx, promptType)
		if err != nil {
			return fmt.Errorf("failed to create prompt type: %w", err)
		}
		fmt.Printf("  > Created\n")
		return nil
	}

	hash := md5.Sum(existingPromptType.Config)
	existingConfigCheckSum := hex.EncodeToString(hash[:])

	hashNew := md5.Sum(promptType.Config)
	newConfigCheckSum := hex.EncodeToString(hashNew[:])

	if (existingPromptType.Description == promptType.Description) && (existingConfigCheckSum == newConfigCheckSum) {
		fmt.Printf("  > ✅ Prompt Type already exists and synced nothing to do\n")
		m.count.Noop++
		return nil
	}

	m.count.OutOfSync++
	if existingPromptType.Description != promptType.Description {
		fmt.Printf("  > 🆘 Prompt Type description out of sync... updating it\n")
		fmt.Printf("     * %s\n", diff(existingPromptType.Description, promptType.Description))
	}

	if existingConfigCheckSum != newConfigCheckSum {
		fmt.Printf("  > 🆘 Prompt Type config out of sync... updating it\n")
		fmt.Printf("     * %s\n", diff(existingConfigCheckSum, newConfigCheckSum))
	}

	if !m.execute {
		return nil
	}

	existingPromptType.Description = promptType.Description
	existingPromptType.Config = promptType.Config

	if err = m.db.UpdatePromptType(ctx, existingPromptType); err != nil {
		return fmt.Errorf("failed to create message type: %w", err)
	}
	fmt.Println("  > ✅ Prompt Type updated")
	return nil
}

func (m *PromptTypeSyncer) selectOrg() (*models.Organization, error) {
	orgNames := make([]string, len(m.orgs)+1)
	for i, org := range m.orgs {
		orgNames[i] = org.Name
	}

	choice := promptui.Select{
		Label: "Select an Organization",
		Items: orgNames,
	}
	idx, _, err := choice.Run()
	if err != nil {
		return nil, err
	}
	if idx == 0 {
		return nil, nil
	}
	return m.orgs[idx], nil
}

type promptOptions struct {
	Validate        promptui.ValidateFunc
	IsConfirm       bool
	PromptTemplates *promptui.PromptTemplates
	Default         string
}

func diff(existing, new string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(existing, new, false)
	return dmp.DiffPrettyText(diffs)
}

func jsonEq(expected, actual json.RawMessage) bool {
	var expectedJSONAsInterface, actualJSONAsInterface interface{}

	if err := json.Unmarshal(expected, &expectedJSONAsInterface); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(actual, &actualJSONAsInterface); err != nil {
		panic(err)
	}

	return assert.ObjectsAreEqual(expectedJSONAsInterface, actualJSONAsInterface)
}

// promptConfirm is just like [prompt] but enforce `IsConfirm` and returns a boolean which is either
// `true` for yes answer or `false` for a no answer.
func promptConfirm(label string, opts *promptOptions) (bool, error) {
	if opts == nil {
		opts = &promptOptions{}
	}

	opts.IsConfirm = true
	transform := func(in string) (bool, error) {
		in = strings.ToLower(in)
		return in == "y" || in == "yes", nil
	}

	return promptT(label, transform, opts)
}

var confirmPromptRegex = regexp.MustCompile("(y|Y|n|N|No|Yes|YES|NO)")

func prompt(label string, opts *promptOptions) (string, error) {
	var templates *promptui.PromptTemplates

	if opts != nil {
		templates = opts.PromptTemplates
	}

	if templates == nil {
		templates = &promptui.PromptTemplates{
			Success: `{{ . | faint }}{{ ":" | faint}} `,
		}
	}

	if opts != nil && opts.IsConfirm {
		// We have no differences
		templates.Valid = `{{ "?" | blue}} {{ . | bold }} {{ "[y/N]" | faint}} `
		templates.Invalid = templates.Valid
	}

	prompt := promptui.Prompt{
		Label:     label,
		Templates: templates,
		Default:   opts.Default,
	}
	if opts != nil && opts.Validate != nil {
		prompt.Validate = opts.Validate
	}

	if opts != nil && opts.IsConfirm {
		prompt.Validate = func(in string) error {
			if !confirmPromptRegex.MatchString(in) {
				return errors.New("answer with y/yes/Yes or n/no/No")
			}

			return nil
		}
	}

	choice, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) {
			// We received Ctrl-C, users wants to abort, nothing else to do, quit immediately
			os.Exit(1)
		}

		if prompt.IsConfirm && errors.Is(err, promptui.ErrAbort) {
			return "false", nil
		}

		return "", fmt.Errorf("running prompt: %w", err)
	}

	return choice, nil
}

// promptT is just like [prompt] but accepts a transformer that transform the `string` into the generic type T.
func promptT[T any](label string, transformer func(string) (T, error), opts *promptOptions) (T, error) {
	choice, err := prompt(label, opts)
	if err == nil {
		return transformer(choice)
	}

	var empty T
	return empty, err
}
