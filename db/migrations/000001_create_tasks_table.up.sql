CREATE TABLE IF NOT EXISTS tasks (
    id serial PRIMARY KEY,
    name VARCHAR ( 100 ) NOT NULL,
    description VARCHAR ( 1000 ),
    status BOOLEAN NOT NULL
);
