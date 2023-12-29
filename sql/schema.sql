-- sqlfluff:templater:raw
-- Create "users" table
CREATE TABLE users (
    id text NOT NULL,
    fullname text NOT NULL DEFAULT '',
    email text NOT NULL,
    email_verified boolean DEFAULT 0,
    active_practice_plan_id TEXT,
    active_practice_plan_started INTEGER,
    PRIMARY KEY (id),
    CHECK (email_verified IN (0, 1)),
    CONSTRAINT plan FOREIGN KEY (active_practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "users_email" to table: "users"
CREATE UNIQUE INDEX users_email ON users (email);

-- Create "credentials" table
CREATE TABLE credentials (
    credential_id blob NOT NULL,
    public_key blob NOT NULL,
    transport blob NOT NULL,
    attestation_type text NOT NULL,
    flags blob NOT NULL,
    authenticator blob NOT NULL,
    user_id text NOT NULL,
    PRIMARY KEY (credential_id),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "credentials_credential_id" to table: "credentials"
CREATE UNIQUE INDEX credentials_credential_id ON credentials (credential_id);

CREATE TABLE pieces (
    id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    composer TEXT,
    measures INTEGER,
    beats_per_measure INTEGER,
    goal_tempo INTEGER,
    user_id TEXT NOT NULL,
    last_practiced INTEGER,
    stage TEXT NOT NULL DEFAULT 'active',
    PRIMARY KEY (id),
    CHECK (LENGTH(title) > 0),
    CHECK (stage IN ('active', 'completed', 'future')),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);


CREATE INDEX pieces_user_id ON pieces (user_id);
CREATE INDEX pieces_user_id_title ON pieces (user_id, title);

CREATE TABLE spots (
    id TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    name TEXT NOT NULL,
    idx INTEGER NOT NULL,
    stage TEXT NOT NULL DEFAULT 'repeat',
    measures TEXT,
    audio_prompt_url TEXT NOT NULL DEFAULT '',
    image_prompt_url TEXT NOT NULL DEFAULT '',
    notes_prompt TEXT NOT NULL DEFAULT '',
    text_prompt TEXT NOT NULL DEFAULT '',
    current_tempo INTEGER,
    last_practiced INTEGER,
    stage_started INTEGER,
    skip_days INTEGER NOT NULL DEFAULT 0,
    priority INTEGER NOT NULL DEFAULT 0,
    CHECK(stage IN ('repeat', 'extra_repeat', 'random', 'interleave', 'interleave_days', 'completed')),
    CHECK(LENGTH(name) > 0),
    CHECK(priority > -3),
    CHECK(priority < 3),
    PRIMARY KEY (id),
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX spots_piece_id ON spots (piece_id);

CREATE TABLE practice_sessions(
    id TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    date INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX practice_sessions_date ON practice_sessions (date);

CREATE TABLE practice_piece (
    practice_session_id TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    measures TEXT NOT NULL,
    PRIMARY KEY (practice_session_id, piece_id),
    CONSTRAINT practice_session FOREIGN KEY (practice_session_id) REFERENCES practice_sessions (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE practice_spot (
    practice_session_id TEXT NOT NULL,
    spot_id TEXT NOT NULL,
    reps INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (practice_session_id, spot_id),
    CONSTRAINT practice_session FOREIGN KEY (practice_session_id) REFERENCES practice_sessions (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT spot FOREIGN KEY (spot_id) REFERENCES spots (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);


CREATE TABLE practice_plans (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    intensity TEXT NOT NULL,
    date INTEGER NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    practice_session_id TEXT,
    CHECK (intensity IN ('light', 'medium', 'heavy')),
    PRIMARY KEY (id),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT ps FOREIGN KEY (practice_session_id) REFERENCES practice_sessions (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);

CREATE TABLE practice_plan_spots (
    practice_plan_id TEXT NOT NULL,
    spot_id TEXT NOT NULL,
    practice_type TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days')),
    PRIMARY KEY (practice_plan_id, spot_id),
    CONSTRAINT plan FOREIGN KEY (practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT spot FOREIGN KEY (spot_id) REFERENCES spots (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);


CREATE TABLE practice_plan_pieces (
    practice_plan_id TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    practice_type TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    CHECK (practice_type IN ('random_spots', 'starting_point')),
    PRIMARY KEY (practice_plan_id, piece_id, practice_type),
    CONSTRAINT plan FOREIGN KEY (practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

