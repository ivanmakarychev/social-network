docker-compose exec db1 mysql -uroot -p65hh0x21GJmlqM \
  -e "GRANT SELECT ON performance_schema.* TO 'social-network-user'@'%';" \
  -e "GRANT GROUP_REPLICATION_ADMIN ON *.* TO 'social-network-user'@'%';" \
  -e "SET @@GLOBAL.group_replication_bootstrap_group=1;" \
  -e "create user 'repl'@'%';" \
  -e "GRANT REPLICATION SLAVE ON *.* TO repl@'%';" \
  -e "GRANT CONNECTION_ADMIN ON *.* TO repl@'%';" \
  -e "GRANT BACKUP_ADMIN ON *.* TO repl@'%';" \
  -e "GRANT GROUP_REPLICATION_STREAM ON *.* TO repl@'%';" \
  -e "flush privileges;" \
  -e "change master to master_user='root' for channel 'group_replication_recovery';" \
  -e "START GROUP_REPLICATION;" \
  -e "SET @@GLOBAL.group_replication_bootstrap_group=0;" \
  -e "SELECT * FROM performance_schema.replication_group_members;"

for N in 2 3
do docker-compose exec db$N mysql -uroot -p65hh0x21GJmlqM \
  -e "change master to master_user='repl' for channel 'group_replication_recovery';" \
  -e "START GROUP_REPLICATION;"
done
