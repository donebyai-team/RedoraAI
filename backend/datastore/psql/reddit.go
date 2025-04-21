package psql

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"keyword/create_keyword.sql",
		"keyword/query_keyword_by_project.sql",
		"sub_reddit/query_sub_reddit_by_filter.sql",
		"sub_reddit/create_sub_reddit.sql",
		"sub_reddit/query_sub_reddit_by_url.sql",
		"sub_reddit/delete_sub_reddit_by_id.sql",
		"sub_reddit/query_sub_reddit_by_id.sql",
		"sub_reddit/query_sub_reddit_by_project.sql",

		"reddit_leads/create_reddit_lead.sql",
		"reddit_leads/query_reddit_lead_by_filter.sql",
		"reddit_leads/query_reddit_lead_by_post_id.sql",
		"reddit_leads/query_reddit_lead_by_status.sql",
		"reddit_leads/update_reddit_lead_status.sql",
		"reddit_leads/query_reddit_lead_by_id.sql",
	})
}

func (r *Database) CreateKeyword(ctx context.Context, keywords *models.Keyword) (*models.Keyword, error) {
	stmt := r.mustGetStmt("keyword/create_keyword.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"keyword":    keywords.Keyword,
		"project_id": keywords.ProjectID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	keywords.ID = id
	return keywords, nil
}

func (r *Database) GetKeywords(ctx context.Context, orgID string) ([]*models.Keyword, error) {
	return getMany[models.Keyword](ctx, r, "keyword/query_keyword_by_project.sql", map[string]any{
		"project_id": orgID,
	})
}

func (r *Database) AddSubReddit(ctx context.Context, subreddit *models.SubReddit) (*models.SubReddit, error) {
	stmt := r.mustGetStmt("sub_reddit/create_sub_reddit.sql")

	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"subreddit_id":         subreddit.SubRedditID,
		"url":                  subreddit.URL,
		"name":                 subreddit.Name,
		"description":          subreddit.Description,
		"project_id":           subreddit.ProjectID,
		"subreddit_created_at": subreddit.SubredditCreatedAt,
		"subscribers":          subreddit.Subscribers,
		"title":                subreddit.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert subreddit: %w", err)
	}

	subreddit.ID = id
	return subreddit, nil
}

func (r *Database) GetSubReddits(ctx context.Context) ([]*models.AugmentedSubReddit, error) {
	subReddits, err := getMany[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_filter.sql", map[string]any{})

	if err != nil {
		return nil, fmt.Errorf("failed to get customer cases: %w", err)
	}
	var results []*models.AugmentedSubReddit
	for _, subreddit := range subReddits {
		keywords, err := r.GetKeywords(ctx, subreddit.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get keywords for project %q: %w", subreddit.ProjectID, err)
		}

		project, err := r.GetProject(ctx, subreddit.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %q: %w", subreddit.ProjectID, err)
		}

		results = append(results, &models.AugmentedSubReddit{
			SubReddit: subreddit,
			Keywords:  keywords,
			Project:   project,
		})
	}

	return results, nil
}

func (r *Database) GetSubRedditsByProject(ctx context.Context, projectID string) ([]*models.SubReddit, error) {
	return getMany[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_project.sql", map[string]any{
		"project_id": projectID,
	})
}

func (r *Database) GetSubRedditByUrl(ctx context.Context, url, projectID string) (*models.SubReddit, error) {
	return getOne[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_url.sql", map[string]any{
		"url":        url,
		"project_id": projectID,
	})
}

func (r *Database) GetSubRedditByID(ctx context.Context, ID string) (*models.SubReddit, error) {
	return getOne[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_id.sql", map[string]any{
		"id": ID,
	})
}

func (r *Database) DeleteSubRedditByID(ctx context.Context, ID string) (*models.SubReddit, error) {
	stmt := r.mustGetStmt("sub_reddit/delete_sub_reddit_by_id.sql")

	var subreddit models.SubReddit
	err := stmt.QueryRowxContext(ctx, map[string]interface{}{
		"id": ID,
	}).StructScan(&subreddit)
	if err != nil {
		return nil, fmt.Errorf("failed to delete subreddit with ID %s: %w", ID, err)
	}

	return &subreddit, nil
}

func (r *Database) GetRedditLeadByPostID(ctx context.Context, projectID, postID string) (*models.RedditLead, error) {
	return getOne[models.RedditLead](ctx, r, "reddit_leads/query_reddit_lead_by_post_id.sql", map[string]any{
		"post_id":    postID,
		"project_id": projectID,
	})
}

func (r *Database) GetRedditLeadByID(ctx context.Context, projectID, id string) (*models.RedditLead, error) {
	return getOne[models.RedditLead](ctx, r, "reddit_leads/query_reddit_lead_by_id.sql", map[string]any{
		"id":         id,
		"project_id": projectID,
	})
}

func (r *Database) UpdateRedditLeadStatus(ctx context.Context, lead *models.RedditLead) error {
	stmt := r.mustGetStmt("reddit_leads/update_reddit_lead_status.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"status":     lead.Status,
		"project_id": lead.ProjectID,
		"id":         lead.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update lead status %q: %w", lead.ID, err)
	}
	return nil
}

func (r *Database) GetRedditLeadsByStatus(ctx context.Context, projectID, status string) ([]*models.RedditLead, error) {
	return getMany[models.RedditLead](ctx, r, "reddit_leads/query_reddit_lead_by_status.sql", map[string]any{
		"status":     status,
		"project_id": projectID,
	})
}

func (r *Database) GetRedditLeadsByRelevancy(ctx context.Context, projectID, relevancy float64, subReddits []string) ([]*models.RedditLead, error) {
	return getMany[models.RedditLead](ctx, r, "reddit_leads/query_reddit_lead_by_filter.sql", map[string]any{
		"subreddit_id":    pq.Array(subReddits),
		"relevancy_score": relevancy,
		"status":          models.RedditLeadStatusNEW,
		"project_id":      projectID,
	})
}

func (r *Database) GetRedditLeadByCommentID(ctx context.Context, projectID, commentID string) (*models.RedditLead, error) {
	panic("implement me")
}

// TODO: Move it under a transaction
func (r *Database) CreateRedditLeads(ctx context.Context, redditLeads []*models.RedditLead) error {
	for _, reddit := range redditLeads {
		stmt := r.mustGetStmt("reddit_leads/create_reddit_lead.sql")
		var id string
		err := stmt.GetContext(ctx, &id, map[string]interface{}{
			"project_id":      reddit.ProjectID,
			"subreddit_id":    reddit.SubRedditID,
			"author":          reddit.Author,
			"post_id":         reddit.PostID,
			"type":            reddit.Type,
			"relevancy_score": reddit.RelevancyScore,
			"post_created_at": reddit.PostCreatedAt,
			"metadata":        reddit.RedditLeadMetadata,
			"description":     reddit.Description,
			"title":           reddit.Title,
		})
		if err != nil {
			return fmt.Errorf("failed to insert reddit_lead post_id [%s]: %w", reddit.PostID, err)
		}
	}
	return nil
}
