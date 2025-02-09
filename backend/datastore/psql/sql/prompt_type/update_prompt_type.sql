UPDATE prompt_types
SET
    description = :description,
    config = :config
WHERE name = :name;