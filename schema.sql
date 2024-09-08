CREATE TABLE administration (vk_id INT PRIMARY KEY NOT NULL);

CREATE TABLE roles_groups (
    name TEXT PRIMARY KEY NOT NULL,
    shown_name TEXT NOT NULL,
    [order] INT NOT NULL
);

CREATE TABLE roles (
    name TEXT PRIMARY KEY NOT NULL,
    hashtag TEXT UNIQUE NOT NULL,
    shown_name TEXT NOT NULL,
    accusative_name TEXT NOT NULL,
    caption_name TEXT NOT NULL,
    [group] TEXT REFERENCES roles_groups(name),
    [order] INT,
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
    status TEXT CHECK(status IN ('Active', 'Freeze')) NOT NULL DEFAULT 'Active'
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
    -- json array for urls
    greeting TEXT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ongoing_polls (
    role TEXT REFERENCES roles(name) PRIMARY KEY NOT NULL,
    post INT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE poll_answer_cache (
    poll_id INT,
    answer_id INT,
    value INT,
    PRIMARY KEY (poll_id, answer_id)
);

-- views
CREATE VIEW reservations_details AS
SELECT
    reservations.vk_id,
    reservations.introduction,
    reservations.deadline,
    reservations.greeting,
    CASE
        WHEN ongoing_polls.post IS NOT NULL THEN 'Poll'
        ELSE reservations.status
    END AS status,
    ongoing_polls.post as poll,
    roles.*
FROM
    reservations
    INNER JOIN roles ON reservations.role = roles.name
    LEFT JOIN ongoing_polls USING(role);

CREATE VIEW polls AS
SELECT
    reservations.role,
    count(*) as count,
    group_concat(vk_id) OVER (
        ORDER BY
            vk_id
    ) AS participants,
    json_group_object(cast(vk_id as text), json(greeting)) AS greetings
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

CREATE VIEW polls_details AS
SELECT
    count,
    participants,
    greetings,
    roles.*,
    ongoing_polls.post
FROM
    polls
    INNER JOIN roles ON polls.role = roles.name
    LEFT JOIN ongoing_polls USING(role);

CREATE VIEW pending_polls AS
SELECT
    *
FROM
    polls_details
WHERE
    name NOT IN (
        SELECT
            role
        FROM
            ongoing_polls
    );

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
            polls
        UNION
        SELECT
            role
        FROM
            members
    );

CREATE VIEW deadlines AS
SELECT
    SUM(diff) as deadline
FROM
    deadline
GROUP BY
    member;

CREATE VIEW members_details AS
SELECT
    members.id,
    members.vk_id,
    members.status,
    deadlines.deadline,
    roles.*
FROM
    members
    INNER JOIN roles ON pending_polls.role = roles.name
    LEFT JOIN deadlines ON members.id = deadline.member;