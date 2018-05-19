package mq

import (
	"sync"

	"errors"

	"fmt"

	"github.com/assembla/cony"
	"github.com/micro/protobuf/proto"
	"github.com/streadway/amqp"
)

var (
	ErrInvalidCallback = errors.New("invalid callback")
	ErrPublisherClosed = errors.New("publish closed")
)

type Delivery struct {
	amqp.Delivery
	Proto proto.Message
}

func (d *Delivery) Ack() error {
	return d.Delivery.Ack(false)
}

type ClientCon struct {
	con *cony.Client

	publishers []*ConPublisher

	closed  chan struct{}
	wg, wg2 sync.WaitGroup

	opts []cony.ClientOpt
	ecb  ErrorHandler
}

func DialAddress(url string, opts ...ClientConOpt) *ClientCon {
	return Dial(WithOptions(cony.URL(url), cony.Backoff(cony.DefaultBackoff)))
}

func Dial(opts ...ClientConOpt) *ClientCon {
	ccon := &ClientCon{closed: make(chan struct{}, 0)}
	for _, o := range opts {
		o(ccon)
	}

	ccon.con = cony.NewClient(ccon.opts...)

	go ccon.loop()
	return ccon
}

func (c *ClientCon) loop() {
	c.wg2.Add(1)
	defer c.wg2.Done()

	for c.con.Loop() {
		select {
		case err := <-c.con.Errors():
			if err != nil && err != (*amqp.Error)(nil) {
				c.handleError(fmt.Errorf("err: %v", err))
			}
		}
	}
}

func (c *ClientCon) errorLoop(ec <-chan error) {
	c.wg.Add(1)
	defer c.wg.Done()

LOOP:
	for {
		select {
		case err := <-ec:
			if err != nil {
				c.handleError(err)
			}
		case _, ok := <-c.closed:
			if !ok {
				break LOOP
			}
		}
	}

	for len(ec) > 0 {
		if err := <-ec; err != nil {
			c.handleError(err)
		}
	}
}

func (c *ClientCon) handleError(err error) {
	fmt.Printf("mqclientcon %v", err)
	if c.ecb != nil {
		c.ecb(err)
	}
}

func (c *ClientCon) Subscribe(opts ...ConSubscribeOpt) *ConSubscribe {
	s := &ConSubscribe{c: c}
	for _, o := range opts {
		o(s)
	}

	go s.run()
	return s
}

func (c *ClientCon) Publish(opts ...ConPublisherOpt) *ConPublisher {
	p := &ConPublisher{c: c, closed: make(chan struct{}, 0)}
	for _, o := range opts {
		o(p)
	}

	if p.q == nil {
		p.q = make(chan *Publishing, 0)
	}

	c.publishers = append(c.publishers, p)

	go p.run()
	return p
}

func (c *ClientCon) Close() {
	// check channel is close
	select {
	case _, closed := <-c.closed:
		if !closed {
			return
		}
	default:
	}

	for _, p := range c.publishers {
		p.Close()
	}

	close(c.closed)
	c.wg.Wait()
	c.con.Close()
	c.wg2.Wait()
}

type ConSubscribe struct {
	c *ClientCon

	q        *cony.Queue
	b        *cony.Binding
	consumer *cony.Consumer
	ack      bool
	cb       []ConSubscribeHandler
}

func (s *ConSubscribe) run() {
	s.c.wg.Add(1)
	defer s.c.wg.Done()

	s.consumer = cony.NewConsumer(s.q)
	s.c.con.Consume(s.consumer)

LOOP:
	for {
		select {
		case msg := <-s.consumer.Deliveries():
			var gerr error

			d := &Delivery{Delivery: msg}
			for _, ccb := range s.cb {
				if gerr = ccb(d); gerr != nil {
					break
				}
			}

			if gerr != nil {
				s.c.handleError(fmt.Errorf("subscribe %v", gerr))
			}
		case err := <-s.consumer.Errors():
			if err != nil {
				s.c.handleError(fmt.Errorf("consumer %v", err))
			}
		case _, closed := <-s.c.closed:
			if !closed {
				break LOOP
			}
		}
	}
}

type Publishing struct {
	amqp.Publishing
	Proto proto.Message
	Key   string
}

type ConPublisher struct {
	c *ClientCon

	e *cony.Exchange
	p *cony.Publisher

	q chan *Publishing
	f []ConPublisherFilter

	wg     sync.WaitGroup
	closed chan struct{}
}

func (p *ConPublisher) run() {
	p.c.wg.Add(1)
	p.wg.Add(1)
	defer p.c.wg.Done()
	defer p.wg.Done()

	fn := func(msg *Publishing) {
		var gerr error

		for _, f := range p.f {
			if gerr = f(msg); gerr != nil {
				break
			}
		}

		if gerr != nil {
			p.c.handleError(fmt.Errorf("filter %v", gerr))
			return
		}

		if err := p.p.PublishWithRoutingKey(msg.Publishing, msg.Key); err != nil {
			p.c.handleError(fmt.Errorf("publish %v", err))
		}
	}

	p.p = cony.NewPublisher(p.e.Name, "")
	p.c.con.Publish(p.p)

LOOP:
	for {
		select {
		case msg := <-p.q:
			fn(msg)
		case _, closed := <-p.closed:
			if !closed {
				break LOOP
			}
		case _, closed := <-p.c.closed:
			if !closed {
				goto EXIT
			}
		}
	}

	for len(p.q) > 0 {
		fn(<-p.q)
	}

	close(p.q)
EXIT:
}

func (p *ConPublisher) publish(pb *Publishing) error {
	if p.IsClosed() {
		return ErrPublisherClosed
	}

	p.q <- pb
	return nil
}

func (p *ConPublisher) PublishData(d []byte) error {
	return p.publish(&Publishing{Publishing: amqp.Publishing{Body: d}})
}

func (p *ConPublisher) PublishKeyData(key string, d []byte) error {
	return p.publish(&Publishing{Key: key, Publishing: amqp.Publishing{Body: d}})
}

func (p *ConPublisher) PublishProto(m proto.Message) error {
	return p.publish(&Publishing{Proto: m})
}

func (p *ConPublisher) PublishKeyProto(key string, m proto.Message) error {
	return p.publish(&Publishing{Key: key, Proto: m})
}

func (p *ConPublisher) IsClosed() bool {
	// check channel is close
	select {
	case _, closed := <-p.closed:
		if !closed {
			return true
		}

		return false
	default:
		return false
	}
}

func (p *ConPublisher) Close() {
	// check channel is close
	if p.IsClosed() {
		return
	}

	close(p.closed)
	p.wg.Wait()
}
