SELECT st.*
FROM subreddit_trackers st
         JOIN subreddits s ON st.subreddit_id = s.id
         JOIN keywords k ON st.keyword_id = k.id
WHERE
    st.last_tracked_at < NOW() - INTERVAL '24 hours'
  AND (s.deleted_at IS NULL AND k.deleted_at IS NULL);
