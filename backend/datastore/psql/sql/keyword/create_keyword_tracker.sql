INSERT INTO keyword_trackers (keyword_id, source_id, project_id)
VALUES (:keyword_id, :source_id, :project_id) RETURNING id;
