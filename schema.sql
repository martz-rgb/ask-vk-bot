CREATE TABLE IF NOT EXISTS administration (vk_id INT PRIMARY KEY NOT NULL);

CREATE TABLE IF NOT EXISTS roles (
    name TEXT PRIMARY KEY NOT NULL,
    tag TEXT NOT NULL,
    shown_name TEXT NOT NULL,
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
    diff INT NOT NULL DEFAULT 0,
    cause TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_points_person ON points(person);

CREATE TABLE IF NOT EXISTS members (
    -- integer primary key -> alias to rowid
    id INTEGER PRIMARY KEY NOT NULL,
    person INT REFERENCES persons(vk_id) NOT NULL,
    role TEXT REFERENCES roles(name) NOT NULL,
    status TEXT CHECK(status IN ('Active', 'Freeze')) NOT NULL DEFAULT 'Active',
		timezone INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS deadline (
    member INT REFERENCES members(id) NOT NULL,
    -- unix time in seconds!
    diff INT NOT NULL DEFAULT 0,
    type TEXT CHECK(
        type IN (
            'Init',
            'Answer',
            'Delay',
            'Rest',
            'Freeze',
            'Other'
        )
    ) NOT NULL DEFAULT 'Other',
    cause TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS init_member_deadline 
AFTER INSERT ON members
	BEGIN
		INSERT INTO deadline(member, diff, type, cause) 
		VALUES (new.id, 
						unixepoch(‘now’, ‘start of day’, ‘+1 day’, ‘-1 second’),
						‘Init’,
						‘init member deadline’);
	END;