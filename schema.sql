USE todo;
CREATE TABLE IF NOT EXISTS todos (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    completed BOOLEAN,
    sortOrder INT
);
INSERT INTO todos VALUES (
    NULL, 
    'Set up MySQL',
    1, 
    0
);
INSERT INTO todos VALUES (
    NULL, 
    'Add more todos',
    0, 
    0
);