-- +goose Up
CREATE TABLE IF NOT EXISTS node_registration (
  hostname TEXT NOT NULL,
  key_id TEXT NOT NULL,
  proto BLOB
);

CREATE TABLE IF NOT EXISTS server_registration (
  hostname TEXT NOT NULL,
  key_id TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_node_registration_hostname ON node_registration (hostname);
CREATE UNIQUE INDEX idx_server_registration_hostname ON server_registration (hostname);

CREATE TABLE IF NOT EXISTS command_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hostname TEXT NOT NULL,
  time DATETIME NOT NULL,   
  success INTEGER,
  output BLOB,
  FOREIGN KEY(hostname) REFERENCES node_registration(hostname)
);

-- +goose Down
DROP TABLE command_records;
DROP TABLE node_registration;
DROP TABLE server_registration;

