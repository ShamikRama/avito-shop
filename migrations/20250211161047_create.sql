-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,   -- Логин для аутентификации
                       password_hash VARCHAR(255) NOT NULL,     -- Хеш пароля
                       balance INTEGER NOT NULL DEFAULT 1000,   -- Текущий баланс (стартовые 1000 монет)
                       CHECK (balance >= 0)                     -- Запрет отрицательного баланса
);
CREATE TABLE items (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) UNIQUE NOT NULL,  -- Название товара (например, "Футболка Avito")
                       price INTEGER NOT NULL              -- Стоимость в монетах
);
CREATE TABLE transfers (
                           id SERIAL PRIMARY KEY,
                           from_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,  -- Отправитель
                           to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,    -- Получатель
                           amount INTEGER NOT NULL,                                      -- Сумма перевода
                           CHECK (amount > 0),                                           -- Запрет нулевых/отрицательных сумм
                           CHECK (from_user_id != to_user_id)                            -- Запрет перевода самому себе
);
CREATE TABLE purchases (
                           id SERIAL PRIMARY KEY,
                           user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,   -- Покупатель
                           item_id INTEGER REFERENCES items(id) ON DELETE CASCADE,   -- Купленный товар
                           quantity INTEGER NOT NULL DEFAULT 1,                      -- Количество товара
                           CHECK (quantity > 0),                                      -- Запрет нулевых/отрицательных значений
                           CONSTRAINT uniq_user_item UNIQUE (user_id, item_id)       -- альтернатива индексу

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
