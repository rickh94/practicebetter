-- sqlfluff:templater:raw
-- Create "users" table
CREATE TABLE users (
    id text NOT NULL,
    name text NULL,
    email text NOT NULL,
    PRIMARY KEY (id)
);
-- Create index "users_email" to table: "users"
CREATE UNIQUE INDEX users_email ON users (email);

-- Create "credentials" table
CREATE TABLE credentials (
    id text NOT NULL,
    credential_id blob NOT NULL,
    public_key blob NOT NULL,
    transport text NOT NULL,
    attetestation_type text NULL,
    flags text NULL,
    authenticator blob NULL,
    user_id text NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT user FOREIGN KEY (user_id) REFERENCES users (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "credentials_credential_id" to table: "credentials"
CREATE UNIQUE INDEX credentials_credential_id ON credentials (credential_id);

-- Create "composers" table
CREATE TABLE composers (
    id text NOT NULL,
    name text NOT NULL,
    PRIMARY KEY (id)
);
-- Create index "composers_name" to table: "composers"
CREATE UNIQUE INDEX composers_name ON composers (name);

-- Create "boxes" table
CREATE TABLE boxes (
    id text NOT NULL,
    name text NOT NULL,
    number integer NOT NULL,
    location text NULL,
    is_full boolean NULL DEFAULT 0,
    PRIMARY KEY (id),
    CHECK (is_full IN (0, 1))
);
-- Create index "box_number" to table: "boxes"
CREATE UNIQUE INDEX box_number ON boxes (number);
-- Create index "box_name" to table: "boxes"
CREATE UNIQUE INDEX box_name ON boxes (name);

-- Create "pieces" table
CREATE TABLE pieces (
    id text NOT NULL,
    title text NOT NULL,
    composer_id text NULL,
    catalog_number text NULL,
    PRIMARY KEY (id),
    CONSTRAINT composer FOREIGN KEY (composer_id) REFERENCES composers (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "piece_title" to table: "pieces"
CREATE INDEX piece_title ON pieces (title);
-- Create index "catalog_number" to table: "pieces"
CREATE INDEX catalog_number ON pieces (catalog_number);

-- Create "items" table
CREATE TABLE items (
    id text NOT NULL,
    description text NOT NULL,
    box_id text NULL,
    instrument text NOT NULL,
    notes text NULL,
    publisher text NULL,
    PRIMARY KEY (id),
    CONSTRAINT box FOREIGN KEY (box_id) REFERENCES boxes (
        id
    ) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "item_box" to table: "items"
CREATE INDEX item_box ON items (box_id);
CREATE INDEX item_description ON items (description);
CREATE INDEX item_instrument ON items (instrument);
CREATE INDEX item_publisher ON items (publisher);

-- Create "items_to_pieces" table
CREATE TABLE items_to_pieces (
    item_id text NOT NULL,
    piece_id text NOT NULL,
    PRIMARY KEY (item_id, piece_id),
    CONSTRAINT item FOREIGN KEY (item_id) REFERENCES items (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT piece FOREIGN KEY (piece_id) REFERENCES pieces (
        id
    ) ON UPDATE NO ACTION ON DELETE CASCADE
);
