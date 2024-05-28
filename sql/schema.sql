-- sqlfluff:templater:raw
-- Create "users" table
CREATE TABLE users (
    id TEXT NOT NULL,
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
    key_id INTEGER,
    mode_id INTEGER,
    PRIMARY KEY (id),
    CHECK (LENGTH(title) > 0),
    CHECK (stage IN ('active', 'completed', 'future')),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT key FOREIGN KEY (key_id) REFERENCES scale_keys (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT mode FOREIGN KEY (mode_id) REFERENCES scale_modes (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);


CREATE INDEX pieces_user_id ON pieces (user_id);
CREATE INDEX pieces_user_id_title ON pieces (user_id, title);

CREATE TABLE spots (
    id TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    name TEXT NOT NULL,
    stage TEXT NOT NULL DEFAULT 'repeat',
    measures TEXT,
    audio_prompt_url TEXT NOT NULL DEFAULT '',
    image_prompt_url TEXT NOT NULL DEFAULT '',
    notes_prompt TEXT NOT NULL DEFAULT '',
    text_prompt TEXT NOT NULL DEFAULT '',
    current_tempo INTEGER,
    last_practiced INTEGER,
    stage_started INTEGER,
    skip_days INTEGER NOT NULL DEFAULT 1,
    priority INTEGER NOT NULL DEFAULT 0,
    section_id TEXT,
    CHECK(stage IN ('repeat', 'extra_repeat', 'random', 'interleave', 'interleave_days', 'completed')),
    CHECK(LENGTH(name) > 0),
    CHECK(priority > -3),
    CHECK(priority < 3),
    CHECK(skip_days > 0),
    PRIMARY KEY (id),
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT section FOREIGN KEY (section_id) REFERENCES sections (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);

CREATE INDEX spots_piece_id ON spots (piece_id);
CREATE INDEX spots_piece_stage ON spots(piece_id, stage);
CREATE INDEX spots_name ON spots (name);

CREATE TABLE practice_plans (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    intensity TEXT NOT NULL,
    date INTEGER NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    practice_notes TEXT,
    last_practiced INTEGER,
    CHECK (intensity IN ('light', 'medium', 'heavy')),
    PRIMARY KEY (id),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX practice_plans_user_id ON practice_plans (user_id);

CREATE TABLE practice_plan_spots (
    practice_plan_id TEXT NOT NULL,
    spot_id TEXT NOT NULL,
    practice_type TEXT NOT NULL,
    evaluation TEXT,
    completed BOOLEAN NOT NULL DEFAULT 0,
    idx INTEGER NOT NULL DEFAULT 0,
    CHECK (practice_type IN ('new', 'extra_repeat', 'interleave', 'interleave_days')),
    CHECK (evaluation = NULL OR evaluation IN ('poor', 'fine', 'excellent')),
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
    sessions INTEGER NOT NULL DEFAULT 1,
    idx INTEGER NOT NULL DEFAULT 0,
    CHECK (practice_type IN ('random_spots', 'starting_point')),
    PRIMARY KEY (practice_plan_id, piece_id, practice_type),
    CONSTRAINT plan FOREIGN KEY (practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE scale_keys (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
    cof INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE scale_modes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    basic BOOLEAN NOT NULL,
    cof INTEGER NOT NULL DEFAULT 0,
    CHECK (name IN ("Major (Ionian)", "Minor (Aeolian)", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Locrian"))
);


CREATE TABLE scales (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id INTEGER NOT NULL,
    mode_id INTEGER NOT NULL,
    FOREIGN KEY (key_id) REFERENCES scale_keys(id),
    FOREIGN KEY (mode_id) REFERENCES scale_modes(id)
);


CREATE TABLE user_scales (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    scale_id INTEGER NOT NULL,
    practice_notes TEXT NOT NULL,
    last_practiced INTEGER,
    reference TEXT NOT NULL,
    working BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (scale_id) REFERENCES scales(id)
);

CREATE TABLE practice_plan_scales (
    practice_plan_id TEXT NOT NULL,
    user_scale_id TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    idx INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (practice_plan_id, user_scale_id),
    CONSTRAINT plan FOREIGN KEY (practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT scale FOREIGN KEY (user_scale_id) REFERENCES user_scales (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE sections (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    FOREIGN KEY (piece_id) REFERENCES pieces(id)
);

CREATE TABLE spots_sections (
    spot_id TEXT NOT NULL,
    section_id TEXT NOT NULL,
    piece_id TEXT NOT NULL,
    PRIMARY KEY (spot_id, section_id),
    FOREIGN KEY (spot_id) REFERENCES spots(id),
    FOREIGN KEY (section_id) REFERENCES sections(id)
);

CREATE TABLE reading (
    id TEXT NOT NULL,
    title TEXT NOT NULL,
    info TEXT,
    completed BOOLEAN NOT NULL DEFAULT 0,
    composer TEXT,
    user_id TEXT NOT NULL,
    PRIMARY KEY (id),
    CHECK (LENGTH(title) > 0),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);


CREATE TABLE practice_plan_reading (
    practice_plan_id TEXT NOT NULL,
    reading_id TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    idx INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (practice_plan_id, reading_id),
    CONSTRAINT plan FOREIGN KEY (practice_plan_id) REFERENCES practice_plans (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT scale FOREIGN KEY (reading_id) REFERENCES reading (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);
