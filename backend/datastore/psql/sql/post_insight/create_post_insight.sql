INSERT INTO post_insights (
    project_id,
    post_id,
    source_id,
    keyword_id,
    relevancy_score,
    topic,
    sentiment,
    highlights,
    metadata)
VALUES (
           :project_id,
           :post_id,
           :source_id,
           :keyword_id,
           :relevancy_score,
           :topic,
           :sentiment,
           :highlights,
           :metadata)
    RETURNING id;