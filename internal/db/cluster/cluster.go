package cluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// NewFromConfig creates a new Cluster from the passed config.DBConf
func NewFromConfig(conf config.DBConf) *Cluster {
	c := newCluster(len(conf.Hosts))
	c.conf = &conf
	c.startReconnector()
	c.AddNodes()
	log.Debug("Created db cluster")
	return c
}

func newCluster(size int) *Cluster {
	c := &Cluster{
		active: make(chan *node, size),
		down:   make(chan *node, size),
		stop:   make(chan interface{}),
	}
	return c
}

// Cluster is a type for holding a db cluster
type Cluster struct {
	active chan *node
	down   chan *node
	stop   chan interface{}
	conf   *config.DBConf
}

type node struct {
	db     *sqlx.DB
	host   string
	active bool
	// lock   sync.RWMutex
}

func (n *node) close() {
	if n.db != nil {
		err := n.db.Close()
		log.Errorf("%s", errorfmt.Full(err))
		n.db = nil
	}
}

// AddNodes adds the nodes specified for this Cluster to the cluster
func (c *Cluster) AddNodes() {
	for _, host := range c.conf.Hosts {
		if err := c.AddNode(host); err != nil {
			log.Errorf("%s", errorfmt.Full(err))
		}
	}
}

// AddNode adds the passed host a a db node to the cluster
func (c *Cluster) AddNode(host string) error {
	log.WithField("host", host).Debug("Adding node to db cluster")
	return c.addNode(
		&node{
			host: host,
		},
	)
}

func (c *Cluster) addNode(n *node) error {
	n.close()
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true", c.conf.User, c.conf.GetPassword(), "tcp", n.host, c.conf.DB)
	db, err := connectDSN(dsn)
	if err != nil {
		n.active = false
		c.down <- n
		log.WithField("dsn", dsn).Debug("Could not connect node")
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
				log.Debug("Stopping re-connector")
				return
			default:
				log.Debug("Run checkNodesDown")
				if c.checkNodesDown() {
					log.Debug("Stopping re-connector")
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
	_ = c.addNode(n)
	for i := 0; i < l; i++ { // check the reminding nodes
		n = <-c.down
		_ = c.addNode(n)
	}
	return false
}

// Close closes the cluster
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
		return nil, errors.WithStack(err)
	}
	db.SetConnMaxLifetime(time.Minute * 4)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

// Transact does a database transaction for the passed function
func (c *Cluster) Transact(rlog log.Ext1FieldLogger, fn func(*sqlx.Tx) error) error {
	for {
		n := c.next(rlog)
		if n == nil {
			return errors.New("no db node available")
		}
		closed, err := n.transact(rlog, fn)
		if !closed {
			return err
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		n.active = false
	}
}

func (n *node) transact(rlog log.Ext1FieldLogger, fn func(*sqlx.Tx) error) (bool, error) {
	err := n.trans(rlog, fn)
	if err != nil {
		e := errorfmt.Error(err)
		switch {
		case e == "sql: database is closed",
			strings.HasPrefix(e, "dial tcp"),
			strings.HasSuffix(e, "closing bad idle connection: EOF"):
			rlog.WithField("host", n.host).Error("Node is down")
			return true, err
		}
	}
	return false, err
}
func (n *node) trans(rlog log.Ext1FieldLogger, fn func(*sqlx.Tx) error) error {
	tx, err := n.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	err = fn(tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			rlog.Errorf("%s", errorfmt.Full(e))
		}
		return err
	}
	return errors.WithStack(tx.Commit())
}

func (c *Cluster) next(rlog log.Ext1FieldLogger) *node {
	rlog.Trace("Selecting a node")
	select {
	case n := <-c.active:
		if n.active {
			c.active <- n
			rlog.WithField("host", n.host).Trace("Selected active node")
			return n
		}
		rlog.WithField("host", n.host).Trace("Found inactive node")
		go c.addNode(n) // try to add node again, if it does not work, will add to down nodes
		return c.next(rlog)
	default:
		rlog.Debug("No active nodes")
		return nil
	}
}

// RunWithinTransaction runs the passed function using the passed transaction; if nil is passed as tx a new transaction
// is created. This is basically a wrapper function, that works with a possible nil-tx
func (c *Cluster) RunWithinTransaction(rlog log.Ext1FieldLogger, tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	if tx == nil {
		return c.Transact(rlog, fn)
	} else {
		return fn(tx)
	}
}
