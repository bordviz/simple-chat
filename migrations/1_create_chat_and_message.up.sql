CREATE TABLE IF NOT EXISTS chat 
(
    id SERIAL PRIMARY KEY UNIQUE,
    first_user_id INTEGER NOT NULL,
    second_user_id INTEGER NOT NULL,
    last_message TEXT,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_chat_first_user_id ON chat(first_user_id);
CREATE INDEX IF NOT EXISTS idx_chat_second_user_id ON chat(second_user_id);
CREATE INDEX IF NOT EXISTS idx_chat_id ON chat(id);

CREATE TABLE IF NOT EXISTS message
(
    id SERIAL PRIMARY KEY UNIQUE,
    chat_id INTEGER NOT NULL REFERENCES chat(id),
    sender INTEGER NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_message_chat_id ON message(chat_id);
CREATE INDEX IF NOT EXISTS idx_message_id ON message(id);