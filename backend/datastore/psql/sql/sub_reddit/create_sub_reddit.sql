INSERT INTO sub_reddits (subreddit_id, url, name, description, organization_id,
subreddit_created_at, last_tracked_at, subscribers, title, updated_at)
VALUES (:subreddit_id, :url, :name, :description, :organization_id,
:subreddit_created_at, :last_tracked_at, :subscribers, :title, :updated_at)
RETURNING id;
