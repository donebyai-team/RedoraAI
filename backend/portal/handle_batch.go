package portal

import (
	"bytes"
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/csv"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"go.uber.org/zap"
	"io"
)

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

	var customerSessions []*services.CreateCustomerSession
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

		promptType, err := p.db.GetPromptTypeByName(ctx, row.ProductType, orgDetails.ID)
		if err != nil {
			return nil, fmt.Errorf("product type[%s], not configured for the given org: %w", row.ProductType, err)
		}

		ppi := &services.CreateCustomerSession{
			FirstName:  row.FirstName,
			LastName:   row.LastName,
			Phone:      row.Phone,
			OrgID:      orgDetails.ID,
			PromptType: promptType.Name,
			DueDate:    row.DueDate,
		}

		customerSessions = append(customerSessions, ppi)
	}

	if csvParseError {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("we ran into an issue, please try again : %w", parseError))
	}

	var count = 0
	rejectedRows := []string{}
	for _, ppi := range customerSessions {
		err := p.customerSessionService.Create(ctx, ppi)
		if err != nil {
			p.logger.Warn("unable to create customerSession", zap.Error(err), zap.String("phone", ppi.Phone))
			rejectedRows = append(rejectedRows, ppi.Phone)
		} else {
			count++
		}
	}

	return response(&pbportal.BatchResp{
		Rows:          int32(len(customerSessions)),
		RowsExtracted: int32(count),
		RejectedRows:  rejectedRows,
	}), nil
}
