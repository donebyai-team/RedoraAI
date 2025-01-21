INSERT INTO organizations (name, feature_flags, tms_platform)
VALUES (:name, :feature_flags, :tms_platform) RETURNING id;