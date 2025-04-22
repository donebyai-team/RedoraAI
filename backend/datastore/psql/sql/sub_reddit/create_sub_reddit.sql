INSERT INTO sub_reddits (subreddit_id, name, description, project_id, subreddit_created_at, title)
VALUES (:subreddit_id, :name, :description, :project_id, :subreddit_created_at, :title)
RETURNING id;
