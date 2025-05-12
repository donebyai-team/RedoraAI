UPDATE keywords set deleted_at = CURRENT_TIMESTAMP WHERE id = :id AND project_id=:project_id;
