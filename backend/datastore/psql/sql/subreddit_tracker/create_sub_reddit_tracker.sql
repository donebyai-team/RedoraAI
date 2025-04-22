INSERT INTO subreddit_trackers (
    subreddit_id,
    keyword_id
)
VALUES (
    :subreddit_id,
    :keyword_id
)
ON CONFLICT (subreddit_id, keyword_id)
DO UPDATE SET
    updated_at = NOW(),
    last_tracked_at = :last_tracked_at,
    newest_tracked_post = :newest_tracked_post,
    oldest_tracked_post = :oldest_tracked_post
    RETURNING id;
