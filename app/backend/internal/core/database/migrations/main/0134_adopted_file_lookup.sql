-- Individual adoption checks one path rather than scanning every library file.
CREATE INDEX files_active_path ON files(lower(relative_path))
WHERE state IN ('present','missing');
