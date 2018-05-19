package mq

import (
	"log"
	"sync"

	"reflect"

	"github.com/assembla/cony"
	"github.com/micro/protobuf/proto"
	"github.com/streadway/amqp"
)

type Publisher struct {
	rw sync.Mutex

	cli        *Client
	exchange   string
	routingKey string

	protoChan      chan proto.Message
	jsonChan       chan interface{}
	keyMessageChan chan *KeyMessage
}

type Consumer struct {
	cli  *Client
	name string
}

func NewClient(url string) *Client {
	return Open(url)
}

func (c *Client) NewPublisher(exchange, routingKey string) *Publisher {
	return &Publisher{
		cli:        c,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (c *Publisher) initProtoWriteExchange() {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.protoChan != nil {
		return
	}

	c.protoChan = make(chan proto.Message, 0)
	go c.cli.WriteExchange(c.exchange, c.routingKey, c.protoChan)
}

func (c *Publisher) initKeyMessageWriteExchange() {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.keyMessageChan != nil {
		return
	}

	c.keyMessageChan = make(chan *KeyMessage, 0)
	go c.cli.WriteExchangeWithKey(c.exchange, c.keyMessageChan)
}

func (c *Publisher) initJsonWriteExchange() {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.jsonChan != nil {
		return
	}

	c.jsonChan = make(chan interface{}, 0)
	go c.cli.WriteExchangeWithJsonMessage(c.exchange, c.routingKey, c.jsonChan)
}

func (c *Publisher) Publish(msg proto.Message) {
	c.initProtoWriteExchange()
	c.protoChan <- msg
}

func (c *Publisher) PublishWithKey(msg *KeyMessage) {
	c.initKeyMessageWriteExchange()
	c.keyMessageChan <- msg
}

func (c *Publisher) PublishJsonMessage(msg interface{}) {
	c.initJsonWriteExchange()
	c.jsonChan <- msg
}

func (c *Client) NewConsumer(name string) *Consumer {
	return &Consumer{cli: c, name: name}
}

func (c *Consumer) Subscribe(f func(amqp.Delivery) error) {
	go c.cli.ReadQueue(&cony.Queue{Name: c.name}, f)
}

func (c *Consumer) SubscribeWithUnmarshal(f interface{}) {
	go c.Subscribe(WithUnmarshal(f))
}

func (c *Consumer) SubscribeExchange(f func(amqp.Delivery) error) {
	go c.cli.ReadExchange(cony.Exchange{Name: c.name}, f)
}

func (c *Consumer) SubscribeExchangeWithUnmarshal(f interface{}) {
	go c.SubscribeExchange(WithUnmarshal(f))
}

func WithUnmarshal(f interface{}) func(amqp.Delivery) error {
	var (
		typerr   = reflect.TypeOf((*error)(nil)).Elem()
		typproto = reflect.TypeOf((*proto.Message)(nil)).Elem()
	)

	// step 1 check function is valid

	if f == nil {
		log.Fatalf("WithCallback : f is nil")
	}

	rfunc := reflect.TypeOf(f)
	if rfunc.Kind() != reflect.Func {
		log.Fatalf("WithCallback : f is not a function")
	}

	if rfunc.NumIn() != 1 {
		log.Fatalf("WithCallback : f parameter size is not 1")
	}

	if rfunc.NumOut() != 1 {
		log.Fatalf("WithCallback : f return size is not 1")
	}

	ityp := rfunc.In(0)
	if ityp.Kind() != reflect.Ptr {
		log.Fatalf("WithCallback : f parameter is not a pointer")
	}

	if ityp.Elem().Kind() != reflect.Struct {
		log.Fatalf("WithCallback : f parameter is not a struct pointer")
	}

	if !ityp.Implements(typproto) {
		log.Fatalf("WithCallback : f parameter is not implements proto.Message")
	}

	otyp := rfunc.Out(0)
	if otyp != typerr {
		log.Fatalf("WithCallback : f return type is not error")
	}

	vfunc := reflect.ValueOf(f)
	return func(delivery amqp.Delivery) error {
		val := reflect.New(ityp.Elem())
		if err := proto.Unmarshal(delivery.Body, val.Interface().(proto.Message)); err != nil {
			return err
		}

		rt := vfunc.Call([]reflect.Value{val})[0]
		if rt.IsNil() {
			return nil
		}

		return rt.Interface().(error)
	}
}
