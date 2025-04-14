INSERT INTO keywords (keyword, project_id)
VALUES (:keyword, :project_id) RETURNING id;
