INSERT INTO sources (external_id, name, description, project_id, metadata, source_type)
VALUES (:external_id, :name, :description, :project_id, :metadata, :source_type)
RETURNING id;
