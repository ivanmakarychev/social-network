docker-compose exec db1 mysql -uroot -p65hh0x21GJmlqM \
  -e "create user 'repl_tarantool'@'%' IDENTIFIED BY '123456789';" \
  -e "GRANT SELECT ON *.* TO repl_tarantool@'%';" \
  -e "GRANT REPLICATION CLIENT ON *.* TO repl_tarantool@'%';" \
  -e "GRANT REPLICATION SLAVE ON *.* TO repl_tarantool@'%';" \
  -e "FLUSH PRIVILEGES;"
