CREATE EXTENSION IF NOT EXISTS postgres_fdw;

CREATE SERVER IF NOT EXISTS shard1 FOREIGN DATA WRAPPER postgres_fdw
    OPTIONS (dbname 'dialogue', host 'dialogue1');

CREATE FOREIGN TABLE IF NOT EXISTS dialogue_message_1 PARTITION OF dialogue_message
    FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1;

CREATE USER MAPPING
    FOR PUBLIC
    SERVER shard1
    OPTIONS (user 'dialogue-user', password 'pG7mDXwwLcfq');