SELECT *
FROM sub_reddits_leads
WHERE project_id = :project_id
  AND relevancy_score >= :relevancy_score
  AND (cardinality(:subreddit_ids) = 0 OR subreddit_id = ANY(:subreddit_ids))
  AND status = :status
ORDER BY created_at DESC;