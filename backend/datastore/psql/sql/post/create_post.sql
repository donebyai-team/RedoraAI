INSERT INTO posts (project_id, title, description, source_id, status, metadata, reason, reference_id)
VALUES (:project_id, :title, :description, :source_id, :status, :metadata, :reason, :reference_id) RETURNING id;
