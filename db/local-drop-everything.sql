/* First drop the database owned by our user. */
DROP DATABASE go_micro_ex_db;

/* Next, drop any privileges owned by our user. */
DROP OWNED BY go_micro_ex_user;

/* Last, drop our user. */
DROP USER go_micro_ex_user;
