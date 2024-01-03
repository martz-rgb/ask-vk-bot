CREATE TABLE IF NOT EXISTS roles (
    name TEXT PRIMARY KEY NOT NULL,
    tag TEXT,
    shown_name TEXT,
    caption_name TEXT,
    album_link TEXT,
    board_link TEXT
);

CREATE TABLE IF NOT EXISTS persons (
    vk_id INT PRIMARY KEY NOT NULL,
    name TEXT,
    points INT
);