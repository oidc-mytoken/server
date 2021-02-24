package cluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

func NewFromConfig(conf config.DBConf) *Cluster {
	c := New(len(conf.Hosts))
	c.AddNodes(conf)
	c.conf = &conf
	log.Debug("Created db cluster")
	return c
}

func New(size int) *Cluster {
	c := &Cluster{
		active: make(chan *node, size),
		down:   make(chan *node, size),
		stop:   make(chan interface{}),
	}
	c.startReconnector()
	return c
}

type Cluster struct {
	active chan *node
	down   chan *node
	stop   chan interface{}
	conf   *config.DBConf
}

type node struct {
	db     *sqlx.DB
	dsn    string
	active bool
	// lock   sync.RWMutex
}

func (n *node) close() {
	if n.db != nil {
		n.db.Close()
		n.db = nil
	}
}

func (c *Cluster) AddNodes(conf config.DBConf) {
	for _, host := range conf.Hosts {
		if err := c.AddNode(conf, host); err != nil {
			log.WithError(err).Error()
		}
	}
}

func (c *Cluster) AddNode(conf config.DBConf, host string) error {
	log.WithField("host", host).Debug("Adding node to db cluster")
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s", conf.User, conf.Password, "tcp", host, conf.DB)
	return c.addNode(&node{
		dsn: dsn,
	})
}

func (c *Cluster) addNode(n *node) error {
	n.close()
	db, err := connectDSN(n.dsn)
	if err != nil {
		n.active = false
		c.down <- n
		log.WithField("dsn", n.dsn).Debug("Could not connect node")
		return err
	}
	n.db = db
	n.active = true
	c.active <- n
	return nil
}

func (c *Cluster) startReconnector() {
	go func() {
		for {
			select {
			case <-c.stop:
				log.Debug("Stopping reconnector")
				return
			default:
				log.Debug("Run checkNodesDown")
				if c.checkNodesDown() {
					log.Debug("Stopping reconnector")
					return
				}
				conf := c.conf
				if conf == nil {
					conf = &config.Get().DB
				}
				time.Sleep(time.Duration(conf.ReconnectInterval) * time.Second)
			}
		}
	}()
}

func (c *Cluster) checkNodesDown() bool {
	var n *node
	select {
	case <-c.stop:
		return true
	case n = <-c.down: // blocks until at least one node is down
		break
	}
	l := len(c.down)
	c.addNode(n)
	for i := 0; i < l; i++ { // check the reminding nodes
		n = <-c.down
		c.addNode(n)
	}
	return false
}

func (c *Cluster) Close() {
	c.stop <- struct{}{}
	for {
		select {
		case active := <-c.active:
			active.close()
		case inactive := <-c.down:
			inactive.close()
		default:
			return
		}
	}
}

func connectDSN(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 4)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

// Transact does a database transaction for the passed function
func (c *Cluster) Transact(fn func(*sqlx.Tx) error) error {
	for {
		n := c.next()
		if n == nil {
			return fmt.Errorf("no db node available")
		}
		closed, err := n.transact(fn)
		if !closed {
			return err
		}
		log.WithError(err).Error()
		n.active = false
	}
}

func (n *node) transact(fn func(*sqlx.Tx) error) (bool, error) {
	err := n.trans(fn)
	if err != nil {
		e := err.Error()
		switch {
		case e == "sql: database is closed",
			strings.HasPrefix(e, "dial tcp"),
			strings.HasSuffix(e, "closing bad idle connection: EOF"):
			log.WithField("dsn", n.dsn).Error("Node is down")
			return true, err
		}
	}
	return false, err
}
func (n *node) trans(fn func(*sqlx.Tx) error) error {
	tx, err := n.db.Beginx()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.WithError(e).Error()
		}
		return err
	}
	return tx.Commit()
}

func (c *Cluster) next() *node {
	log.Debug("Selecting a node")
	select {
	case n := <-c.active:
		if n.active {
			c.active <- n
			log.WithField("dsn", n.dsn).Debug("Selected active node")
			return n
		}
		log.WithField("dsn", n.dsn).Debug("Found inactive node")
		go c.addNode(n) // try to add node again, if it does not work, will add to down nodes
		return c.next()
	default:
		log.Debug("No active nodes")
		return nil
	}
}

// RunWithinTransaction runs the passed function using the passed transaction; if nil is passed as tx a new transaction is created. This is basically a wrapper function, that works with a possible nil-tx
func (c *Cluster) RunWithinTransaction(tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	if tx == nil {
		return c.Transact(fn)
	} else {
		return fn(tx)
	}
}
