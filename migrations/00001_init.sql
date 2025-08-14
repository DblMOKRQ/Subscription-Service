-- +goose Up
-- Создание таблицы подписок
CREATE TABLE subscriptions (
                               id UUID PRIMARY KEY,
                               service_name VARCHAR(255) NOT NULL,
                               price INTEGER NOT NULL CHECK (price > 0),
                               user_id UUID NOT NULL,
                                -- start_date DATE,
                               start_date VARCHAR(7) NOT NULL CHECK (start_date ~ '^\d{2}-\d{4}$'),
    end_date VARCHAR(7) CHECK (end_date ~ '^\d{2}-\d{4}$')
    -- end_date DATE
);

-- Индексы для ускорения поиска
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_start_date ON subscriptions(start_date);
CREATE INDEX idx_subscriptions_end_date ON subscriptions(end_date);
CREATE INDEX idx_subscriptions_user_service ON subscriptions(user_id, service_name);

-- Для поиска по периоду
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);


-- +goose Down
DROP TABLE IF EXISTS subscriptions;
