INSERT INTO posts (project_id, title, description, source_id, status, metadata, reason, reference_id, schedule_at)
VALUES (:project_id, :title, :description, :source_id, :status, :metadata, :reason, :reference_id, :schedule_at) RETURNING id;
