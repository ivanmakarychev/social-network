CREATE TABLE IF NOT EXISTS dialogue_message
(
    message_id        bigserial not null,
    profile_id_1      bigint    not null,
    profile_id_2      bigint    not null,
    profile_id_author bigint    not null,
    ts                timestamp not null default current_timestamp,
    text              text      not null
) PARTITION BY HASH (profile_id_1, profile_id_2);

CREATE TABLE IF NOT EXISTS dialogue_message_1 PARTITION OF dialogue_message
    FOR VALUES WITH (MODULUS 2, REMAINDER 0);

CREATE INDEX IF NOT EXISTS dialogue_message_1_idx ON dialogue_message_1 (profile_id_1, profile_id_2);
