package psql

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	"google.golang.org/grpc/codes"
)

func init() {
	registerFiles([]string{
		"keyword/keyword.sql",
	})
}

func (r *Database) CreateKeyword(ctx context.Context, keywords *models.Keyword) (*models.Keyword, error) {
	stmt := r.mustGetStmt("keyword/keyword.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"keyword":         keywords.Keyword,
		"organization_id": keywords.OrgID,
	})

	if err != nil {
		return nil, connect.NewError(connect.Code(codes.InvalidArgument), fmt.Errorf("failed to create keyword for organization: %w", err))
	}

	keywords.ID = id
	return keywords, nil
}
