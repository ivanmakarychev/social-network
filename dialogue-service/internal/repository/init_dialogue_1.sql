CREATE TABLE IF NOT EXISTS dialogue_message
(
    profile_id_1      bigint    not null,
    profile_id_2      bigint    not null,
    profile_id_author bigint    not null,
    ts                timestamp not null default current_timestamp,
    text              text      not null
) PARTITION BY HASH (profile_id_1, profile_id_2);

ALTER TABLE dialogue_message ADD COLUMN status int not null default 2;

CREATE TABLE IF NOT EXISTS dialogue_message_1 PARTITION OF dialogue_message
    FOR VALUES WITH (MODULUS 2, REMAINDER 0);

CREATE INDEX IF NOT EXISTS dialogue_message_1_idx ON dialogue_message_1 (profile_id_1, profile_id_2, ts);

CREATE EXTENSION IF NOT EXISTS postgres_fdw;

CREATE SERVER IF NOT EXISTS shard2 FOREIGN DATA WRAPPER postgres_fdw
    OPTIONS (dbname 'dialogue', host 'dialogue2');

CREATE FOREIGN TABLE IF NOT EXISTS dialogue_message_2 PARTITION OF dialogue_message
    FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2;

CREATE USER MAPPING IF NOT EXISTS
    FOR PUBLIC
    SERVER shard2
    OPTIONS (user 'dialogue-user', password 'pG7mDXwwLcfq');
