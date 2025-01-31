package csv

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"io"
	"strings"
)

var _ io.Closer = (*BatchCSVReader)(nil)

type BatchCSVReader struct {
	*csv.Reader
	logger *zap.Logger

	rowID     int
	headers   []string
	closer    io.ReadCloser
	validator *validator.Validate
}

// Close implements io.Closer.
func (r *BatchCSVReader) Close() error {
	return r.closer.Close()
}

type BufferCloser struct {
	*bytes.Buffer
}

func (b *BufferCloser) Close() error {
	return nil
}

func NewBatchCSVReaderFromReader(reader io.ReadCloser, logger *zap.Logger, tracer logging.Tracer) (*BatchCSVReader, error) {
	br := bufio.NewReader(reader)
	r, _, err := br.ReadRune()
	if err != nil {
		return nil, fmt.Errorf("reading BOM: %w", err)
	}
	if r != '\uFEFF' {
		br.UnreadRune() // Not a BOM -- put the rune back
	}

	csvReader := csv.NewReader(br)
	csvReader.ReuseRecord = true

	return &BatchCSVReader{
		Reader:    csvReader,
		closer:    reader,
		logger:    logger,
		validator: NewStructValidator(),
	}, nil
}

func (r *BatchCSVReader) readHeader() error {
	re, err := r.Reader.Read()
	if err != nil {
		return err
	}
	// Underlying CSV reader reuses the slice, so we need to copy it and normalize to lowercase
	headers := make([]string, len(re))
	for i, h := range re {
		headers[i] = strings.ToLower(h)
	}

	foundHeaders := 0
	for _, mHeader := range shouldHaveHeaders {
		if utils.Contains(headers, mHeader) {
			foundHeaders++
			break
		}
	}

	if foundHeaders == 0 {
		return fmt.Errorf("missing columns: Either %s should be present", strings.Join(shouldHaveHeaders, ", "))
	}

	r.headers = headers
	r.rowID++
	return nil
}

func (r *BatchCSVReader) Read(ctx context.Context) (*BatchRow, error) {
	defer func() {
		r.rowID++
	}()

	if r.rowID == 0 {
		err := r.readHeader()
		if err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}
	}

	batchRow := NewBatchRow()
	record, err := r.Reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading row %d: %w", r.rowID, err)
	}

	for i, name := range r.headers {
		switch name {
		case ROW_FIRST_NAME:
			batchRow.SetFirstName(record[i])
		case ROW_LAST_NAME:
			batchRow.SetLastName(record[i])
		case ROW_PHONE:
			batchRow.SetPhone(record[i])
		case ROW_DUE_DATE:
			date, _, ok := utils.ParseDateLikeInput(record[i])
			if !ok {
				return nil, fmt.Errorf("row %d: unable to decode due_date %q, accepted formats are YYYY-MM-DD, MM/DD/YYYY and RFC3339", r.rowID, record[i])
			}
			batchRow.SetDueDate(date)
		case ROW_PRODUCT_TYPE:
			batchRow.SetProductType(record[i])
		}
	}

	err = r.Validate(batchRow)
	if err != nil {
		return nil, fmt.Errorf("row %d: %w", r.rowID, err)
	}
	return batchRow, nil
}

func (r *BatchCSVReader) Validate(row *BatchRow) error {
	err := r.validator.Struct(row)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return fmt.Errorf("invalid column value [%s]: [%s]", err.Field(), err.Value())
		}

	}
	return err
}
