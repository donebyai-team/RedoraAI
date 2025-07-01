INSERT INTO post_insights (
    project_id,
    post_id,
    source_type,
    relevancy_score,
    topic,
    sentiment,
    highlights,
    metadata)
VALUES (
           :project_id,
           :post_id,
           :source_type,
           :relevancy_score,
           :topic,
           :sentiment,
           :highlights,
           :metadata)
    RETURNING id;