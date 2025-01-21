UPDATE organizations
SET
    feature_flags = :feature_flags
WHERE id = :id;