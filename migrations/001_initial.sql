CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL
);
CREATE INDEX idx_sessions_token ON sessions(token);

CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    image_path TEXT NOT NULL DEFAULT '',
    prep_time TEXT NOT NULL DEFAULT '',
    cook_time TEXT NOT NULL DEFAULT '',
    servings TEXT NOT NULL DEFAULT '',
    content_md TEXT NOT NULL DEFAULT '',
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE recipe_tags (
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
);

CREATE TABLE user_recipes (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    tried BOOLEAN NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, recipe_id)
);

CREATE TABLE shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(recipe_id, owner_id, shared_with_id)
);

CREATE VIRTUAL TABLE recipes_fts USING fts5(
    title, description, content_md,
    content='recipes', content_rowid='id'
);

CREATE TRIGGER recipes_ai AFTER INSERT ON recipes BEGIN
    INSERT INTO recipes_fts(rowid, title, description, content_md)
    VALUES (new.id, new.title, new.description, new.content_md);
END;

CREATE TRIGGER recipes_ad AFTER DELETE ON recipes BEGIN
    INSERT INTO recipes_fts(recipes_fts, rowid, title, description, content_md)
    VALUES ('delete', old.id, old.title, old.description, old.content_md);
END;

CREATE TRIGGER recipes_au AFTER UPDATE ON recipes BEGIN
    INSERT INTO recipes_fts(recipes_fts, rowid, title, description, content_md)
    VALUES ('delete', old.id, old.title, old.description, old.content_md);
    INSERT INTO recipes_fts(rowid, title, description, content_md)
    VALUES (new.id, new.title, new.description, new.content_md);
END;
