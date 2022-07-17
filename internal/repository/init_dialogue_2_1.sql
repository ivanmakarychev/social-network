CREATE EXTENSION IF NOT EXISTS postgres_fdw;

CREATE SERVER IF NOT EXISTS shard2 FOREIGN DATA WRAPPER postgres_fdw
    OPTIONS (dbname 'dialogue', host 'dialogue2');

CREATE FOREIGN TABLE IF NOT EXISTS dialogue_message_2 PARTITION OF dialogue_message
    FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2;

CREATE USER MAPPING
    FOR PUBLIC
    SERVER shard2
    OPTIONS (user 'dialogue-user', password 'pG7mDXwwLcfq');
