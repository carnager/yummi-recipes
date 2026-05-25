CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

INSERT OR IGNORE INTO settings (key, value) VALUES ('registration_enabled', '1');
INSERT OR IGNORE INTO settings (key, value) VALUES ('llm_classify', '1');
INSERT OR IGNORE INTO settings (key, value) VALUES ('llm_clean_instructions', '1');
