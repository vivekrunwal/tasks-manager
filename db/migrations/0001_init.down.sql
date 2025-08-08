DROP TRIGGER IF EXISTS update_task_modified_trigger ON tasks;
DROP FUNCTION IF EXISTS update_task_modified_column();
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS task_status;
