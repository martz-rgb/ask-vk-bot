CREATE TABLE IF NOT EXISTS administration (
    vk_id INT PRIMARY KEY NOT NULL
);

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
    gallery TEXT,
    birthday TEXT
);

CREATE TABLE IF NOT EXISTS points (
    person INT REFERENCES persons(vk_id) NOT NULL,
    diff INT,
    cause TEXT,
    timestamp DATETIME
);
CREATE INDEX IF NOT EXISTS idx_points_person ON points(person);

CREATE TABLE IF NOT EXISTS members (
    id INT PRIMARY KEY NOT NULL,
    person INT REFERENCES persons(vk_id) NOT NULL,
    role TEXT REFERENCES roles(name) NOT NULL,
    status TEXT CHECK(status IN ('Active','Freeze')) NOT NULL DEFAULT 'Active'
)