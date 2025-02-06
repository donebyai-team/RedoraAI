INSERT INTO organizations (name, feature_flags)
VALUES (:name, :feature_flags) RETURNING id;