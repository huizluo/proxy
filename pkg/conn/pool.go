package conn

import (
	"log"
	"sync"
	"time"
)

type Pool interface {
	Get()(conn interface{},err error)
	Put(conn interface{})
	Len()int
	Remove(conn interface{})
	RemoveAll()
}

type poolConfig struct {
	Factory    func() (interface{}, error)
	IsActive   func(interface{}) bool
	Release    func(interface{})
	InitialCap int
	MaxCap     int
}

type connPool struct {
	connCh chan interface{}
	mu *sync.Mutex
	conf poolConfig
}

func (c connPool) Get() (conn interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case conn = <-c.connCh:
		if c.conf.IsActive(conn){
			return
		}
		c.conf.Release(conn)
	default:
		conn,err = c.conf.Factory()
		if err!=nil{
			return nil,err
		}
		return conn,nil
	}
	return
}

func (c connPool) Put(conn interface{}) {
	if conn == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.conf.IsActive(conn) {
		c.conf.Release(conn)
	}
	select {
	case c.connCh <- conn:
	default:
		c.conf.Release(conn)
	}
}

func (c connPool) Len() int {
	return len(c.connCh)
}

func (c connPool) Remove(conn interface{}) {
	panic("implement me")
}

func (c connPool) RemoveAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.connCh)
	for conn := range c.connCh {
		c.conf.Release(conn)
	}
	c.connCh = make(chan interface{}, c.conf.InitialCap)
}

func NewConnPool(config poolConfig)(pool Pool,err error){
	p := connPool{
		conf: config,
		connCh:  make(chan interface{}, config.MaxCap),
		mu:   &sync.Mutex{},
	}
	//log.Printf("pool MaxCap:%d", poolConfig.MaxCap)
	if config.MaxCap > 0 {
		err = p.initAutoFill(false)
		if err == nil {
			p.initAutoFill(true)
		}
	}
	return &p, nil
}

func (p *connPool) initAutoFill(async bool) (err error) {
	var worker = func() (err error) {
		for {
			//log.Printf("pool fill: %v , len: %d", p.Len() <= p.config.InitialCap/2, p.Len())
			if p.Len() <= p.conf.InitialCap/2 {
				p.mu.Lock()
				errN := 0
				for i := 0; i < p.conf.InitialCap; i++ {
					c, err := p.conf.Factory()
					if err != nil {
						errN++
						if async {
							continue
						} else {
							p.mu.Unlock()
							return err
						}
					}
					select {
					case p.connCh <- c:
					default:
						p.conf.Release(c)
						break
					}
					if p.Len() >= p.conf.InitialCap {
						break
					}
				}
				if errN > 0 {
					log.Printf("fill conn pool fail , ERRN:%d", errN)
				}
				p.mu.Unlock()
			}
			if !async {
				return
			}
			time.Sleep(time.Second * 2)
		}
	}
	if async {
		go worker()
	} else {
		err = worker()
	}
	return

}
