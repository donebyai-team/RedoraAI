UPDATE sources
SET
    metadata = :metadata
WHERE
    id = :id;
