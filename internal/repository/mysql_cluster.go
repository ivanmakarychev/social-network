package repository

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ivanmakarychev/social-network/internal/config"
)

type DBInstance interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	Begin() (*sql.Tx, error)
}

type Cluster interface {
	Master() DBInstance
	Replica() DBInstance
}

type MySQLCluster struct {
	sync.RWMutex
	cfg             config.Database
	conns           map[string]*sql.DB
	master          string
	replicas        []string
	replicaSelector uint64
	hm              HostMapper
}

func NewMySQLCluster(cfg config.Database, hm HostMapper) *MySQLCluster {
	if hm == nil {
		hm = defaultHostMapper
	}
	return &MySQLCluster{
		conns: map[string]*sql.DB{},
		cfg:   cfg,
		hm:    hm,
	}
}

//Init инициализация: создание коннектов и выполнение скриптов
func (c *MySQLCluster) Init() error {
	err := c.Connect()
	if err != nil {
		return err
	}
	return initDB(c.Master())
}

//Connect создание коннектов
func (c *MySQLCluster) Connect() error {
	c.Lock()
	defer c.Unlock()

	var err error
	c.master = strings.Split(c.cfg.Master, ":")[0]
	c.conns[c.master], err = createMySQLConnectionWithRetry(c.cfg.Master, c.cfg, c.hm, 60)
	if err != nil {
		return errors.New("failed to create master: " + err.Error())
	}

	c.replicas = make([]string, 0, len(c.cfg.Replicas))
	for _, replicaAddr := range c.cfg.Replicas {
		var replica *sql.DB
		var err error
		replica, err = createMySQLConnectionWithRetry(replicaAddr, c.cfg, c.hm, 60)
		if err != nil {
			return errors.New("failed to create replica: " + err.Error())
		}
		replicaHost := strings.Split(replicaAddr, ":")[0]
		c.conns[replicaHost] = replica
		c.replicas = append(c.replicas, replicaHost)
	}

	return nil
}

//Close закрывает соединения с базами
func (c *MySQLCluster) Close() {
	c.Lock()
	defer c.Unlock()

	for _, db := range c.conns {
		_ = db.Close()
	}
}

//Master возвращает мастера
func (c *MySQLCluster) Master() DBInstance {
	c.RLock()
	defer c.RUnlock()
	return &dbInstanceMaster{
		db:      c.conns[c.master],
		cluster: c,
	}
}

//Replica возвращает одну из реплик или мастер, если реплик нет
func (c *MySQLCluster) Replica() DBInstance {
	c.RLock()
	defer c.RUnlock()

	replicasCount := len(c.replicas)
	if replicasCount == 0 {
		return c.conns[c.master]
	}
	replicaIdx := atomic.AddUint64(&c.replicaSelector, 1) % uint64(replicasCount)
	return c.conns[c.replicas[int(replicaIdx)]]
}

func (c *MySQLCluster) changeMaster() DBInstance {
	c.Lock()
	defer c.Unlock()

	if len(c.replicas) == 0 {
		log.Println("cannot change master: no replicas")
		return c.conns[c.master]
	}

	host, err := c.getNewMasterHostWithRetries(c.conns[c.replicas[0]])
	if err != nil {
		log.Println("changeMaster: failed to getNewMasterHostWithRetries:", err)
		return c.conns[c.master]
	}
	log.Println("master changed: new master", host)
	c.master = host
	return &dbInstanceMaster{db: c.conns[c.master], cluster: c}
}

type dbInstanceMaster struct {
	db      *sql.DB
	cluster *MySQLCluster
}

func (i *dbInstanceMaster) Begin() (*sql.Tx, error) {
	tx, err := i.db.Begin()
	switch analyseError(err) {
	case errConnectionDead:
		newMaster := i.cluster.changeMaster()
		return newMaster.Begin()
	case errReadOnly:
		for k := 0; k < 60; k++ {
			time.Sleep(time.Second)
			tx, err = i.db.Begin()
			switch analyseError(err) {
			case noError:
				return tx, err
			case errReadOnly:
				continue
			default:
				return tx, err
			}
		}
		return tx, err
	}
	return tx, err
}

func (i *dbInstanceMaster) Query(query string, args ...any) (*sql.Rows, error) {
	rows, err := i.db.Query(query, args...)
	if analyseError(err) == errConnectionDead {
		newMaster := i.cluster.changeMaster()
		return newMaster.Query(query, args...)
	}
	return rows, err
}

func (i *dbInstanceMaster) Exec(query string, args ...any) (sql.Result, error) {
	rs, err := i.db.Exec(query, args...)
	switch analyseError(err) {
	case errConnectionDead:
		newMaster := i.cluster.changeMaster()
		return newMaster.Exec(query, args...)
	case errReadOnly:
		for k := 0; k < 60; k++ {
			time.Sleep(time.Second)
			rs, err = i.db.Exec(query, args...)
			switch analyseError(err) {
			case noError:
				return rs, err
			case errReadOnly:
				continue
			default:
				return rs, err
			}
		}
		return rs, err
	}
	return rs, err
}

type errorKind int

const (
	noError errorKind = iota
	errUnknown
	errConnectionDead
	errReadOnly
)

func analyseError(err error) errorKind {
	if err == nil {
		return noError
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return errConnectionDead
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		if dnsError.IsNotFound {
			return errConnectionDead
		}
	}
	switch err.Error() {
	case "invalid connection", "driver: bad connection":
		return errConnectionDead
	}
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return errUnknown
	}
	const readOnly = uint16(1290)
	if me.Number == readOnly {
		return errReadOnly
	}
	return errUnknown
}

func (c *MySQLCluster) getNewMasterHostWithRetries(db *sql.DB) (string, error) {
	var host string
	var err error
	for i := 0; i < 60; i++ {
		host, err = getMasterHost(db)
		if err != nil {
			if errors.Is(err, errMasterNotFound) {
				time.Sleep(time.Second)
				continue
			}
		}
		masterDB := c.conns[host]
		err = masterDB.Ping()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return host, err
	}
	return host, errMasterNotFound
}

var (
	errMasterNotFound = errors.New("master host not found")
)

func getMasterHost(db *sql.DB) (string, error) {
	const query = "SELECT MEMBER_HOST FROM performance_schema.replication_group_members WHERE MEMBER_ROLE='PRIMARY';"
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var host string
		err = rows.Scan(&host)
		return host, err
	}
	return "", errMasterNotFound
}

type HostMapper func(host string) string

func defaultHostMapper(host string) string {
	return host
}
