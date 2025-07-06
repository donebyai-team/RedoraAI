package portal

import (
	"connectrpc.com/connect"
	"context"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) CreatePost(ctx context.Context, c *connect.Request[pbcore.PostSettings]) (*connect.Response[pbcore.Post], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	newPost := &models.Post{
		ProjectID: project.ID,
		SourceID:  c.Msg.GetSourceId(),
	}

	if c.Msg.ReferenceId != nil {
		newPost.ReferenceID = c.Msg.ReferenceId
	}

	createdPost, err := p.postService.CreatePost(ctx, newPost, c.Msg)

	if err != nil {
		return nil, err
	}

	return connect.NewResponse(new(pbcore.Post).FromModel(createdPost)), nil
}

func (p *Portal) GetPosts(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetPostsResponse], error) {
	//actor, err := p.gethAuthContext(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//
	//project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	//if err != nil {
	//	return nil, err
	//}

	// call service to fetch posts by project id
	return nil, nil
}
