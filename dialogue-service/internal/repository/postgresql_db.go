package repository

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/config"
	"github.com/jackc/pgx/v4"
)

type DialogueDB interface {
	GetConn() *pgx.Conn
}

type ShardedDialogueDB struct {
	shards  []*pgx.Conn
	counter uint64
}

func NewShardedDialogueDB(cfg config.DialogueDatabase, hm HostMapper) (*ShardedDialogueDB, error) {
	if hm == nil {
		hm = defaultHostMapper
	}
	d := ShardedDialogueDB{}
	for _, shardAddr := range cfg.Shards {
		conn, err := createPostgreSQLConnectionWithRetry(
			context.Background(),
			shardAddr,
			cfg,
			hm,
			60,
		)
		if err != nil {
			return nil, err
		}
		d.shards = append(d.shards, conn)
	}
	return &d, nil
}

func (s *ShardedDialogueDB) Init(ctx context.Context) error {
	for i, conn := range s.shards {
		if err := initDialogueDBInstance(
			ctx,
			conn,
			fmt.Sprintf("./internal/repository/init_dialogue_%d.sql", i+1),
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *ShardedDialogueDB) Close() {
	for _, c := range s.shards {
		_ = c.Close(context.Background())
	}
}

func (s *ShardedDialogueDB) GetConn() *pgx.Conn {
	idx := int(atomic.AddUint64(&s.counter, 1)) % len(s.shards)
	return s.shards[idx]
}

func createPostgreSQLConnectionWithRetry(
	ctx context.Context,
	addr string,
	cfg config.DialogueDatabase,
	hm HostMapper,
	retries int,
) (*pgx.Conn, error) {
	addrSplit := strings.Split(addr, ":")
	host := hm(addrSplit[0])

	const dbSourceFmt = "postgres://%s:%s@%s:%s/%s"
	dbSource := fmt.Sprintf(dbSourceFmt, cfg.User, cfg.Password, host, addrSplit[1], cfg.DbName)

	log.Println("trying to open db", dbSource)
	db, err := pgx.Connect(ctx, dbSource)
	for i := 0; err != nil && i < retries; i++ {
		log.Println("retrying opening db")
		db, err = pgx.Connect(ctx, dbSource)
	}
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	for i := 0; err != nil && i < retries; i++ {
		time.Sleep(time.Second)
		log.Println("retrying pinging db")
		err = db.Ping(ctx)
	}
	return db, err
}

func initDialogueDBInstance(ctx context.Context, db *pgx.Conn, scriptFile string) error {
	log.Println("init DB started")

	script, err := os.Open(scriptFile)
	if err != nil {
		return fmt.Errorf("failed to open sql script file: %s", err)
	}
	defer script.Close()

	scanner := bufio.NewScanner(script)

	sb := strings.Builder{}

	counter := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			sb.WriteString(line)
			sb.WriteRune(' ')
		} else {
			continue
		}
		if strings.HasSuffix(line, ";") {
			query := sb.String()
			sb.Reset()
			log.Println("[query]", query)
			_, err = db.Exec(ctx, query)
			if err != nil {
				log.Println("[init db] bad query:", query, "[error]", err)
				return fmt.Errorf("failed to execute sql script file: %s", err)
			}
			counter++
		}
	}

	log.Println("init DB finished.", counter, "queries executed")

	return nil
}

type HostMapper func(host string) string

func defaultHostMapper(host string) string {
	return host
}
