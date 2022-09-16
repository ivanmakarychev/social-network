box.schema.user.passwd('pass');

box.schema.space.create('binlog_pos_space', {id = 512});
box.space.binlog_pos_space:create_index('primary', {
	type = 'HASH',
	parts = {1, 'unsigned'}
});

box.schema.space.create('profile', {
	id = 513,
	field_count = 3
});
box.space.profile:create_index('primary', {
	type = 'HASH',
	parts = {1, 'unsigned'}
});

box.schema.space.create('friends', {
	id = 514,
	field_count = 2
});
box.space.friends:create_index('primary', {
	type = 'HASH',
	parts = {1, 'unsigned'}
});
