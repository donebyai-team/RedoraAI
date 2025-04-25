INSERT INTO sub_reddits_leads (
    project_id,
    author,
    subreddit_id,
    keyword_id,
    post_id,
    type,
    relevancy_score,
    post_created_at,
    metadata,
    title,
    description)
VALUES (
    :project_id,
    :author,
    :subreddit_id,
    :keyword_id,
    :post_id,
    :type,
    :relevancy_score,
    :post_created_at,
    :metadata,
    :title,
    :description)
RETURNING id;