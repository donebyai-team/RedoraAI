package portal

import (
	"connectrpc.com/connect"
	"context"
	"github.com/shank318/doota/datastore"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) GetInsights(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.InsightsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	insights, err := p.db.GetInsights(ctx, project.ID, datastore.LeadsFilter{
		RelevancyScore: 90,
		Limit:          100,
		Offset:         0,
		DateRange:      pbportal.DateRangeFilter_DATE_RANGE_7_DAYS,
	})
	if err != nil {
		return nil, err
	}

	insightsProto := make([]*pbcore.PostInsight, 0, len(insights))
	for _, insight := range insights {
		insightsProto = append(insightsProto, new(pbcore.PostInsight).FromModel(insight))
	}

	return connect.NewResponse(&pbportal.InsightsResponse{Insights: insightsProto}), nil
}
