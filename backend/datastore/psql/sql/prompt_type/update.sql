UPDATE message_types
SET
    description = :description,
    category = :category,
    config = :config
WHERE name = :name;