CREATE TABLE administration (vk_id INT PRIMARY KEY NOT NULL);

CREATE TABLE roles_groups (
    name TEXT PRIMARY KEY NOT NULL,
    shown_name TEXT NOT NULL,
    "order" INT NOT NULL
);

CREATE TABLE roles (
    name TEXT PRIMARY KEY NOT NULL,
    hashtag TEXT UNIQUE NOT NULL,
    shown_name TEXT NOT NULL,
    accusative_name TEXT NOT NULL,
    caption_name TEXT NOT NULL,
    "group" TEXT REFERENCES roles_groups(name),
    album INT,
    board INT
);

CREATE TABLE info (
    vk_id INT PRIMARY KEY NOT NULL,
    gallery TEXT,
    birthday TEXT
);

CREATE TABLE points (
    vk_id INT NOT NULL,
    diff INT NOT NULL DEFAULT 0,
    cause TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_points_vk_id ON points(vk_id);

CREATE TABLE members (
    -- integer primary key -> alias to rowid
    id INTEGER PRIMARY KEY NOT NULL,
    vk_id INT,
    role TEXT REFERENCES roles(name) NOT NULL,
    status TEXT CHECK(status IN ('Active', 'Freeze')) NOT NULL DEFAULT 'Active',
    timezone INT NOT NULL DEFAULT 0
);

CREATE TABLE deadline (
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

CREATE TRIGGER init_member_deadline
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

CREATE TABLE schedule (
    -- alias to rowid
    id INTEGER PRIMARY KEY NOT NULL,
    kind TEXT CHECK (
        kind IN (
            'Polls',
            'Greetings',
            'Answers',
            'Free Answers',
            'Leavings'
        )
    ) NOT NULL,
    query TEXT NOT NULL,
    time_points TEXT NOT NULL
);

CREATE TABLE reservations (
    vk_id INT PRIMARY KEY NOT NULL,
    role TEXT REFERENCES roles(name) NOT NULL,
    status TEXT AS (
        CASE
            WHEN is_confirmed = 0 THEN 'Under Consideration'
            ELSE CASE
                WHEN greeting IS NULL THEN 'In Progress'
                ELSE 'Done'
            END
        END
    ),
    -- id of vk message contained information
    introduction INT NOT NULL,
    is_confirmed INT NOT NULL DEFAULT 0,
    deadline DATETIME,
    -- urls for images 
    greeting TEXT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE polls (
    role TEXT REFERENCES roles(name) PRIMARY KEY NOT NULL,
    post INT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- views
CREATE VIEW reservations_details AS
SELECT
    reservations.vk_id,
    reservations.introduction,
    reservations.deadline,
    reservations.greeting,
    CASE
        WHEN polls.post IS NOT NULL THEN 'Poll'
        ELSE reservations.status
    END AS status,
    polls.post as poll,
    roles.*
FROM
    reservations
    INNER JOIN roles ON reservations.role = roles.name
    LEFT JOIN polls USING(role);

CREATE VIEW pending_polls AS
SELECT
    reservations.role,
    count(*) as count,
    group_concat(vk_id) OVER (
        ORDER BY
            vk_id
    ) AS participants,
    group_concat(greeting, ';') OVER (
        ORDER BY
            vk_id
    ) AS greetings
FROM
    reservations
WHERE
    NOT EXISTS(
        SELECT
            *
        FROM
            reservations AS other
        WHERE
            other.role = reservations.role
            AND other.status != 'Done'
    )
GROUP BY
    role;

CREATE VIEW pending_polls_details AS
SELECT
    count,
    participants,
    greetings,
    roles.*
FROM
    pending_polls
    INNER JOIN roles ON pending_polls.role = roles.name
WHERE
    roles.name NOT IN (
        SELECT
            role
        FROM
            polls
    );

CREATE VIEW ongoing_polls AS
SELECT
    *
FROM
    polls
WHERE
    post IS NOT NULL;

CREATE VIEW available_roles AS
SELECT
    *
FROM
    roles
WHERE
    name NOT IN (
        SELECT
            role
        FROM
            pending_polls
        UNION
        SELECT
            role
        FROM
            members
    );