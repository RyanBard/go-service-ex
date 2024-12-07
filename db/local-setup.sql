CREATE USER go_micro_ex_user PASSWORD 'changeit';
CREATE DATABASE go_micro_ex_db WITH OWNER go_micro_ex_user;

\connect go_micro_ex_db

\i db/initial-schema.sql
\i db/initial-data.sql

-- ALTER ROLE go_micro_ex_ddl SUPERUSER;
-- or maybe: CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- or maybe: CREATE EXTENSION IF NOT EXISTS "pgcrypto";

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO go_micro_ex_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO go_micro_ex_user;

-- if you do this, make sure you run this comment on the owner user or else they won't take affect
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON ALL SEQUENCES TO go_micro_ex_user;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON ALL TABLES TO go_micro_ex_user;
