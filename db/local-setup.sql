CREATE USER go_micro_ex_user PASSWORD 'changeit';
CREATE DATABASE go_micro_ex_db WITH OWNER go_micro_ex_user;

\connect go_micro_ex_db

\i db/initial-schema.sql
\i db/initial-data.sql

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO go_micro_ex_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO go_micro_ex_user;
