package portal

import (
	"bytes"
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/csv"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dhttp"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (p *Portal) HandleBatchAdmin(agent agents.AIAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.Logger(ctx, p.logger)

		logger.Debug("received HandleBatchAdmin request", zap.String("path", r.URL.Path))

		// Parse the multipart form with a 10MB memory buffer
		// Parse the multipart form with a 10MB memory buffer
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		// Retrieve organization_id from the form
		organizationID := r.FormValue("organization_id")
		if organizationID == "" {
			http.Error(w, "Missing organization_id", http.StatusBadRequest)
			return
		}

		// Retrieve the file from the form field named "file"
		file, _, err := r.FormFile("csv")
		if err != nil {
			http.Error(w, "Unable to get file from form-data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the CSV file into a byte slice
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		csvData := buf.Bytes()

		logger.Info("handling HandleBatchAdmin request", zap.String("organizationID", organizationID))
		req := &connect.Request[pbportal.BatchReq]{
			Msg: &pbportal.BatchReq{
				CsvData:        csvData,
				OrganizationId: organizationID,
			},
		}
		resp, err := p.Batch(r.Context(), req)
		if err != nil {
			dhttp.WriteError(r.Context(), w, derr.UnexpectedError(r.Context(), err))
			return
		}

		dhttp.WriteJSON(ctx, w, resp.Msg)
	}
}

func (p *Portal) Batch(ctx context.Context, req *connect.Request[pbportal.BatchReq]) (*connect.Response[pbportal.BatchResp], error) {
	// Read the CSV data
	data := &csv.BufferCloser{Buffer: bytes.NewBuffer(req.Msg.CsvData)}
	reader, err := csv.NewBatchCSVReaderFromReader(data, p.logger, p.tracer)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unable to create CSV data: %w", err))
	}

	// Get org details
	orgDetails, err := p.db.GetOrganizationById(ctx, req.Msg.OrganizationId)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetcg org details: %w", err))
	}

	var customerCases []*services.CreateCustomerCase
	csvParseError := false
	var parseError error
	for {
		row, err := reader.Read(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			csvParseError = true
			parseError = err
			break
		}

		promptType, err := p.db.GetPromptTypeByName(ctx, row.ProductType)
		if err != nil {
			return nil, fmt.Errorf("product type[%s], not configured for the given org: %w", row.ProductType, err)
		}

		ppi := &services.CreateCustomerCase{
			FirstName:  row.FirstName,
			LastName:   row.LastName,
			Phone:      row.Phone,
			OrgID:      orgDetails.ID,
			PromptType: promptType.Name,
			DueDate:    row.DueDate,
		}

		customerCases = append(customerCases, ppi)
	}

	if csvParseError {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("we ran into an issue, please try again : %w", parseError))
	}

	var count = 0
	rejectedRows := []string{}
	for _, ppi := range customerCases {
		err := p.customerCaseService.Create(ctx, ppi)
		if err != nil {
			p.logger.Warn("unable to create customerSession", zap.Error(err), zap.String("phone", ppi.Phone))
			rejectedRows = append(rejectedRows, ppi.Phone)
		} else {
			count++
		}
	}

	return response(&pbportal.BatchResp{
		Rows:          int32(len(customerCases)),
		RowsExtracted: int32(count),
		RejectedRows:  rejectedRows,
	}), nil
}
