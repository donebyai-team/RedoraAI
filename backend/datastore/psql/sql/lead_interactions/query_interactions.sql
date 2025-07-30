SELECT
    li.*,
    l.title AS post_title,
    jsonb_build_object(
            'post_url', l.metadata ->> 'post_url',
            'automated_comment_url', l.metadata ->> 'automated_comment_url',
            'suggested_comment', l.metadata ->> 'suggested_comment',
            'suggested_dm', l.metadata ->> 'suggested_dm',
            'dm_url', l.metadata ->> 'dm_url'
    ) AS lead_metadata
FROM
    lead_interactions li
        JOIN
    leads l ON l.id = li.lead_id
WHERE
    li.project_id = :project_id
  AND (CAST(:start_datetime AS timestamp) IS NULL OR li.created_at >= :start_datetime)
  AND (CAST(:end_datetime AS timestamp) IS NULL OR li.created_at < :end_datetime)
  AND li.schedule_at IS NOT NULL
ORDER BY
    schedule_at DESC
LIMIT :limit;
