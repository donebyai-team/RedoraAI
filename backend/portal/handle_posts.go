package portal

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/shank318/doota/integrations/reddit"
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
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%w", err))
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

	protoPosts := make([]*pbcore.PostDetail, 0, len(posts))
	for _, post := range posts {
		proto := new(pbcore.Post).FromAugmentedModel(post)
		var postID string
		if post.PostID != nil {
			postID = *post.PostID
		}

		augmented := &pbcore.PostDetail{
			Post:       proto,
			SourceName: post.Source.Name,
			PostUrl:    reddit.GetPostURL(postID, post.Source.Name),
		}
		protoPosts = append(protoPosts, augmented)
	}

	return connect.NewResponse(&pbportal.GetPostsResponse{
		Posts: protoPosts,
	}), nil
}

func (p *Portal) UpdatePost(ctx context.Context, c *connect.Request[pbcore.UpdatePostRequest]) (*connect.Response[pbcore.Post], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	if c.Msg.Post == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing post in request"))
	}

	postModel := c.Msg.Post.ToModel()
	postModel.ProjectID = project.ID

	updatedPost, err := p.postService.UpdatePost(ctx, postModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%w", err))
	}

	return connect.NewResponse(new(pbcore.Post).FromModel(updatedPost)), nil
}

func (p *Portal) DeletePost(ctx context.Context, c *connect.Request[pbcore.DeletePostRequest]) (*connect.Response[emptypb.Empty], error) {
	_, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	err = p.postService.DeletePost(ctx, c.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to delete post: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
