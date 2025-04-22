SELECT *
FROM sub_reddits_leads
WHERE project_id = :project_id
  AND relevancy_score >= :relevancy_score
  AND (:subreddit_ids = '{}' OR subreddit_id = ANY(:subreddit_ids))
  AND status = :status
ORDER BY created_at DESC;