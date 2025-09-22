-- Drop indexes
DROP INDEX IF EXISTS idx_files_folder_id;
DROP INDEX IF EXISTS idx_folders_parent_id;
DROP INDEX IF EXISTS idx_folders_user_id;

-- Remove folder_id from files table
ALTER TABLE files DROP COLUMN folder_id;

-- Drop folders table
DROP TABLE IF EXISTS folders;