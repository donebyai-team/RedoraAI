package portal

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Create post using value types
	newPost := &models.Post{
		ProjectID:   project.ID,
		SourceID:    c.Msg.SourceId,
		ID:          c.Msg.GetId(),
		ReferenceID: c.Msg.ReferenceId,
		Metadata: models.PostMetadata{
			Settings: models.PostSettings{
				Topic:       c.Msg.Topic,
				Context:     c.Msg.Context,
				Goal:        c.Msg.Goal,
				Tone:        c.Msg.Tone,
				ReferenceID: c.Msg.ReferenceId,
			},
		},
	}

	if c.Msg.GetReferenceId() != "" {
		newPost.ReferenceID = c.Msg.ReferenceId
	}

	createdPost, err := p.postService.CreatePost(ctx, newPost, project)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(new(pbcore.Post).FromModel(createdPost)), nil
}

func (p *Portal) GetPosts(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetPostsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	posts, err := p.db.GetPostsByProjectID(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	protoPosts := make([]*pbcore.AugmentedPost, 0, len(posts))
	for _, post := range posts {
		proto := new(pbcore.Post).FromAugmentedModel(post)
		augmented := &pbcore.AugmentedPost{
			Post:       proto,
			SourceName: post.Source.Name,
		}
		protoPosts = append(protoPosts, augmented)
	}

	return connect.NewResponse(&pbportal.GetPostsResponse{
		Posts: protoPosts,
	}), nil
}

func (p *Portal) SchedulePost(ctx context.Context, c *connect.Request[pbcore.SchedulePostRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	post, err := p.db.GetPostByID(ctx, c.Msg.Id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.New(codes.NotFound, "Post not found").Err()
	}

	//Validate payload and project ownership
	err = validation.ValidateSchedulePost(c.Msg, post, project.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	scheduleAt := c.Msg.GetScheduleAt().AsTime()

	if err := p.db.SchedulePost(ctx, c.Msg.Id, scheduleAt); err != nil {
		return nil, status.New(codes.Internal, "Failed to schedule post").Err()
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
