package portal

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/streamingfast/logging"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shank318/doota/models"
)

func (p *Portal) Self(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.User], error) {
	logger := logging.Logger(ctx, p.logger)

	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	logger.Info("handling self")

	user, err := p.db.GetUserById(ctx, actor.ID)
	if err != nil {
		return nil, fmt.Errorf("get self: %w", err)
	}

	var organizations []*models.Organization
	if actor.IsPlatformAdmin() {
		organizations, err = p.db.GetOrganizations(ctx)
		if err != nil {
			return nil, fmt.Errorf("get all organizations: %w", err)
		}
	} else {
		org, err := p.db.GetOrganizationById(ctx, actor.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("get organization: %w", err)
		}
		organizations = append(organizations, org)
	}
	sort.SliceStable(organizations, func(i, j int) bool {
		return organizations[i].ID == actor.OrganizationID
	})

	projects, onboardingDone, err := p.getProjects(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	userProto := new(pbportal.User).FromModel(user, organizations)
	userProto.Projects = projects
	userProto.IsOnboardingDone = onboardingDone

	return connect.NewResponse(userProto), nil
}
