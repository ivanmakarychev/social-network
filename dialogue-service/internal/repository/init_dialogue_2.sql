CREATE TABLE IF NOT EXISTS dialogue_message
(
    profile_id_1      bigint    not null,
    profile_id_2      bigint    not null,
    profile_id_author bigint    not null,
    ts                timestamp not null default current_timestamp,
    text              text      not null
) PARTITION BY HASH (profile_id_1, profile_id_2);

CREATE TABLE IF NOT EXISTS dialogue_message_2 PARTITION OF dialogue_message
    FOR VALUES WITH (MODULUS 2, REMAINDER 1);

CREATE INDEX IF NOT EXISTS dialogue_message_2_idx ON dialogue_message_2 (profile_id_1, profile_id_2, ts);

CREATE EXTENSION IF NOT EXISTS postgres_fdw;

CREATE SERVER IF NOT EXISTS shard1 FOREIGN DATA WRAPPER postgres_fdw
    OPTIONS (dbname 'dialogue', host 'dialogue1');

CREATE FOREIGN TABLE IF NOT EXISTS dialogue_message_1 PARTITION OF dialogue_message
    FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1;

CREATE USER MAPPING IF NOT EXISTS
    FOR PUBLIC
    SERVER shard1
    OPTIONS (user 'dialogue-user', password 'pG7mDXwwLcfq');
