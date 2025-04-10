INSERT INTO sub_reddits (subreddit_id, url, name, description, organization_id,
subreddit_created_at, subscribers, title)
VALUES (:subreddit_id, :url, :name, :description, :organization_id,
:subreddit_created_at, :subscribers, :title)
RETURNING id;
