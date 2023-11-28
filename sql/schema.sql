-- sqlfluff:templater:raw
-- Create "users" table
CREATE TABLE users (
    id text NOT NULL,
    fullname text NOT NULL DEFAULT '',
    email text NOT NULL,
    email_verified boolean DEFAULT 0,
    PRIMARY KEY (id),
    CHECK (email_verified IN (0, 1))
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
    PRIMARY KEY (id),
    CHECK (LENGTH(title) > 0),
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
    CHECK(stage IN ('repeat', 'random', 'interleave', 'interleave_days', 'complete')),
    CHECK(LENGTH(name) > 0),
    PRIMARY KEY (id),
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX spots_piece_id ON spots (piece_id);
