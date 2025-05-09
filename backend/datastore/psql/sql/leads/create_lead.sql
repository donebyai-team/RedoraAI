INSERT INTO leads (
    project_id,
    author,
    source_id,
    keyword_id,
    post_id,
    type,
    relevancy_score,
    post_created_at,
    metadata,
    title,
    description,
    intents)
VALUES (
    :project_id,
    :author,
    :source_id,
    :keyword_id,
    :post_id,
    :type,
    :relevancy_score,
    :post_created_at,
    :metadata,
    :title,
    :description,
    :intents)
RETURNING id;