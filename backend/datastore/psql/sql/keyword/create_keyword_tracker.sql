INSERT INTO keyword_trackers (keyword_id, source_id)
VALUES (:keyword_id, :source_id) RETURNING id;
