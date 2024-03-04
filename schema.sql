CREATE TABLE IF NOT EXISTS administration (vk_id INT PRIMARY KEY NOT NULL);

CREATE TABLE IF NOT EXISTS roles (
    name TEXT PRIMARY KEY NOT NULL,
    hashtag TEXT NOT NULL,
    shown_name TEXT NOT NULL,
    accusative_name TEXT,
    caption_name TEXT,
    album_link TEXT,
    board_link TEXT
);

CREATE TABLE IF NOT EXISTS info (
    vk_id INT PRIMARY KEY NOT NULL,
    gallery TEXT,
    birthday TEXT
);

CREATE TABLE IF NOT EXISTS points (
    vk_id INT NOT NULL,
    diff INT NOT NULL DEFAULT 0,
    cause TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_points_vk_id ON points(vk_id);

CREATE TABLE IF NOT EXISTS members (
    -- integer primary key -> alias to rowid
    id INTEGER PRIMARY KEY NOT NULL,
    vk_id INT,
    role TEXT REFERENCES roles(name) NOT NULL,
    status TEXT CHECK(status IN ('Active', 'Freeze')) NOT NULL DEFAULT 'Active',
    timezone INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS deadline (
    member INT REFERENCES members(id) NOT NULL,
    -- unix time in seconds!
    diff INT NOT NULL DEFAULT 0,
    kind TEXT CHECK(
        kind IN (
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
AFTER
INSERT
    ON members BEGIN
INSERT INTO
    deadline(member, diff, kind, cause)
VALUES
    (
        new.id,
        unixepoch(
            'now',
            'start of day',
            '+1 day',
            '-1 second'
        ),
        'Init',
        'init member deadline'
    );

END;

CREATE TABLE IF NOT EXISTS reservations (
    -- alias to rowid
    id INTEGER PRIMARY KEY NOT NULL,
    role TEXT REFERENCES roles(name) NOT NULL,
    -- only one reservation per user
    vk_id INT NOT NULL UNIQUE,
    deadline DATETIME,
    status TEXT CHECK(
        status IN (
            'Under Consideration',
            'In Progress',
            'Done',
            'Poll'
        )
    ) NOT NULL DEFAULT 'Under Consideration',
    -- id of vk message contained information
    info INT NOT NULL,
    -- vk attachments contained greeting (attachments\text)
    greeting TEXT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- views
CREATE VIEW IF NOT EXISTS available_roles AS
SELECT
    *
FROM
    roles
WHERE
    name NOT IN (
        SELECT
            role
        FROM
            members
    )
    AND name NOT IN (
        SELECT
            role AS reserved_role
        FROM
            reservations
        WHERE
            NOT EXISTS(
                SELECT
                    *
                FROM
                    reservations AS other
                where
                    other.role = reservations.role
                    AND (
                        other.status = 'Under Consideration'
                        OR other.status = 'In Progress'
                    )
            )
    );

CREATE VIEW IF NOT EXISTS reservations_details AS
SELECT
    *
FROM
    reservations,
    roles
WHERE
    roles.name = reservations.role;

CREATE VIEW IF NOT EXISTS reservations_polls AS
SELECT
    roles.*,
    group_concat(vk_id) AS participants,
    group_concat(greeting, ';') AS greetings
FROM
    reservations,
    roles
WHERE
    NOT EXISTS(
        SELECT
            *
        FROM
            reservations AS other
        where
            other.role = reservations.role
            AND other.status != 'Done'
    )
    AND reservations.role = roles.name
GROUP BY
    role
ORDER BY
    role;