package ai

import (
	"context"
	"encoding/csv"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type InputRow struct {
	Title              string
	Description        string
	Author             string
	ProductName        string
	ProductDescription string
	TargetPersona      string
	IsRelevant         string
	ActualRelevancy    string
	PublicComment      string
	PublicDM           string
	PostLink           string
}

type OutputRow struct {
	Title         string
	LLMModel      string
	PublicComment string
	PublicDM      string
	Relevancy     float64
	COTRelevancy  string
	COTComment    string
	COTDM         string
}

type ModelStats struct {
	Correct       int
	Total         int
	Errors        int
	TotalDuration time.Duration
}

var llmModels = []string{
	"redora-dev-gpt-4.1-mini-2025-04-14",
	"redora-dev-gpt-4.1-2025-04-14",
	"redora-dev-claude-3-7-sonnet-20250219",
	"redora-dev-gpt-4o-2024-08-06",
}

func TestRedoraCase(t *testing.T) {
	t.Log("Running Redora Case Tests")
	file, err := os.Open("testdata/redora_tests.csv")
	if err != nil {
		log.Fatal("Failed to open CSV:", err)
	}
	defer file.Close()

	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	if err != nil {
		t.FailNow()
	}
	defaultModel := models.LLMModel("redora-dev-gpt-4.1-2025-04-14")
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), defaultModel, LangsmithConfig{}, debugStore)
	if err != nil {
		t.FailNow()
	}

	reader := csv.NewReader(file)
	headers, _ := reader.Read()
	rows, _ := reader.ReadAll()

	modelStats := map[string]*ModelStats{}
	var outputRows []OutputRow

	for index, row := range rows {
		input := parseRow(headers, row)
		isRelevantColumn := strings.ToLower(input.IsRelevant)
		if isRelevantColumn != "yes" && isRelevantColumn != "no" {
			t.Fatal("IsRelevant must be Yes or No")
		}

		isRelevant := isRelevantColumn == "yes"

		for _, model := range llmModels {
			t.Logf("Testing model: %s for row %d", model, index)
			project := &models.Project{
				ID:                 "XXX",
				OrganizationID:     "XXXXX",
				Name:               input.ProductName,
				ProductDescription: input.Description,
				CustomerPersona:    input.TargetPersona,
			}

			post := &models.Lead{
				Author:      input.Author,
				Title:       utils.Ptr(input.Title),
				Description: input.Description,
			}

			org := &models.Organization{
				FeatureFlags: models.OrganizationFeatureFlags{RelevancyLLMModel: models.LLMModel(model)},
			}

			start := time.Now()
			relevant, _, err := ai.IsRedditPostRelevant(context.Background(), org, project, post, logger)
			duration := time.Since(start)

			if _, ok := modelStats[model]; !ok {
				modelStats[model] = &ModelStats{}
			}
			stats := modelStats[model]
			stats.Total++
			stats.TotalDuration += duration

			if err != nil {
				stats.Errors++
				continue
			}

			predictedRelevant := relevant.IsRelevantConfidenceScore >= 0.9
			if predictedRelevant == isRelevant {
				stats.Correct++
			}

			outputRows = append(outputRows, OutputRow{
				Title:         input.Title,
				LLMModel:      model,
				PublicComment: relevant.SuggestedComment,
				PublicDM:      relevant.SuggestedDM,
				Relevancy:     relevant.IsRelevantConfidenceScore,
				COTRelevancy:  relevant.ChainOfThoughtIsRelevant,
				COTComment:    relevant.ChainOfThoughtSuggestedComment,
				COTDM:         relevant.ChainOfThoughtSuggestedDM,
			})
		}
	}

	// Print summary
	t.Log("Model Summary:")
	for model, stats := range modelStats {
		accuracy := float64(stats.Correct) / float64(stats.Total) * 100
		avgTime := stats.TotalDuration / time.Duration(stats.Total)
		t.Logf("Model: %s | Accuracy: %.2f%% | Errors: %d | Avg Response Time: %s",
			model, accuracy, stats.Errors, avgTime)
	}

	// Write output CSV
	writeOutputCSV("testdata/redora_tests_output.csv", outputRows)
}

func parseRow(headers, row []string) InputRow {
	data := map[string]string{}
	for i, h := range headers {
		data[h] = row[i]
	}
	return InputRow{
		Title:              data["Title"],
		Description:        data["Description"],
		Author:             data["Author"],
		ProductName:        data["ProductName"],
		ProductDescription: data["ProductDescription"],
		TargetPersona:      data["TargetCustomerPersona"],
		IsRelevant:         data["IsRelevant"],
		ActualRelevancy:    data["Actual Relevancy"],
		PublicComment:      data["Public_Comment"],
		PublicDM:           data["Public_Dm"],
		PostLink:           data["Post link"],
	}
}

func writeOutputCSV(filename string, rows []OutputRow) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Failed to create output CSV:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Title", "LLMModel", "PublicComment", "PublicDM", "RelevancyScore", "COTRelevancy", "COTComment", "COTDM"})
	for _, row := range rows {
		writer.Write([]string{
			row.Title,
			row.LLMModel,
			row.PublicComment,
			row.PublicDM,
			strconv.FormatFloat(row.Relevancy, 'f', 2, 64),
			row.COTRelevancy,
			row.COTComment,
			row.COTDM,
		})
	}
}
