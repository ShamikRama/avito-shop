-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goose_db_version (
    id SERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance INTEGER NOT NULL,
    CHECK (balance >= 0)
    );

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INTEGER NOT NULL
    );

INSERT INTO items (name, price) VALUES
                                    ('t-shirt', 80),
                                    ('cup', 20),
                                    ('book', 50),
                                    ('pen', 10),
                                    ('powerbank', 200),
                                    ('hoody', 300),
                                    ('umbrella', 200),
                                    ('socks', 10),
                                    ('wallet', 50),
                                    ('pink-hoody', 500);

CREATE TABLE IF NOT EXISTS transfers (
    id SERIAL PRIMARY KEY,
    from_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    CHECK (amount > 0),
    CHECK (from_user_id != to_user_id)
    );

CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    item_id INTEGER REFERENCES items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    CHECK (quantity > 0),
    CONSTRAINT uniq_user_item UNIQUE (user_id, item_id)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS goose_db_version;
-- +goose StatementEnd