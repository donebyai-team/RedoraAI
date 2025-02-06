INSERT INTO prompt_types (
       name,
       description,
       organization_id,
       config
    ) VALUES (
        :name,
        :description,
        :organization_id,
        :config
) RETURNING *;


