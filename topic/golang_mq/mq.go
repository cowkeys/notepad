package mq

import (
	"encoding/json"
	"fmt"
	"sync"

	"encoding/base64"

	"github.com/assembla/cony"
	"github.com/micro/protobuf/proto"
	"github.com/streadway/amqp"
)

type KeyMessage struct {
	Key     string
	Message proto.Message
}

type KeyMessageJson struct {
	Key     string
	Message interface{}
}

type Client struct {
	c     *cony.Client
	run   bool
	close chan bool
	wg    sync.WaitGroup
	errcb func(err error)
}

func Open(url string) *Client {
	c := &Client{
		c: cony.NewClient(
			cony.URL(url),
			cony.Backoff(cony.DefaultBackoff),
		),
		run:   true,
		close: make(chan bool, 0),
	}
	go c.runLoop()
	return c
}

func (c *Client) runLoop() {
	for c.run && c.c.Loop() {
		select {
		case err := <-c.c.Errors():
			if err == nil {
				continue
			}

			c.sendError(err)

			fmt.Printf("mqclient error: %v", err)
		case <-c.close:
			fmt.Printf("mqclient closed")
		}
	}
}

func (c *Client) ReadQueue(q *cony.Queue, f func(amqp.Delivery) error) {
	c.wg.Add(1)
	defer c.wg.Done()

	consumer := cony.NewConsumer(q)
	c.c.Consume(consumer)

	for c.run {
		select {
		case msg := <-consumer.Deliveries():
			if err := f(msg); err != nil {
				b64 := base64.StdEncoding.EncodeToString(msg.Body)
				fmt.Printf("mqclient handler [%v] message: %v err: %v", q.Name, b64, err)
			}

			msg.Ack(false)
		case err := <-consumer.Errors():
			if err == nil {
				continue
			}

			c.sendError(err)

			fmt.Printf("mqclient consumer [%v] err: %v", q.Name, err)
		case <-c.close:
		}
	}

	fmt.Printf("mqclient read end queue: [%v]", q.Name)
}

func (c *Client) writeExchangeMessage(pb *cony.Publisher, kmsg *KeyMessage) error {
	msg := kmsg.Message

	d, err := proto.Marshal(msg)
	if err != nil {
		return err

	}

	err = pb.PublishWithRoutingKey(amqp.Publishing{Body: d}, kmsg.Key)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) writeExchangeJsonMessage(pb *cony.Publisher, kmsg *KeyMessageJson) error {
	msg := kmsg.Message

	d, err := json.Marshal(msg)
	if err != nil {
		return err

	}

	err = pb.PublishWithRoutingKey(amqp.Publishing{Body: d}, kmsg.Key)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) WriteExchange(exchange, routing_key string, ch chan proto.Message) {
	c.wg.Add(1)
	defer c.wg.Done()

	pb := cony.NewPublisher(exchange, "")
	c.c.Publish(pb)

	send := func(msg proto.Message) {
		if err := c.writeExchangeMessage(pb, &KeyMessage{routing_key, msg}); err != nil {
			bmsg, _ := json.Marshal(msg)
			fmt.Printf("mqpublisher write failed msg: %v key: %v err: %v", bmsg, routing_key, err)
		}
	}

	// send message or wait close
	for c.run {
		select {
		case msg := <-ch:
			send(msg)
		case <-c.close:
		}
	}

	// send message immediate
	for len(ch) > 0 {
		send(<-ch)
	}

	// close channel
	close(ch)

	fmt.Printf("mqclient write end exchange: [%v]", exchange)
}

func (c *Client) WriteExchangeWithKey(exchange string, ch chan *KeyMessage) {
	c.wg.Add(1)
	defer c.wg.Done()

	pb := cony.NewPublisher(exchange, "")
	c.c.Publish(pb)

	send := func(msg *KeyMessage) {
		if err := c.writeExchangeMessage(pb, msg); err != nil {
			bmsg, _ := json.Marshal(msg.Message)
			fmt.Printf("mqpublisher write failed msg: %v key: %v err: %v", bmsg, msg.Key, err)
		}
	}

	// send message or wait close
	for c.run {
		select {
		case msg := <-ch:
			send(msg)
		case <-c.close:
		}
	}

	// send message immediate
	for len(ch) > 0 {
		send(<-ch)
	}

	// close channel
	close(ch)

	fmt.Printf("mqclient write end exchange: [%v]", exchange)
}

func (c *Client) WriteExchangeWithJsonMessage(exchange, routing_key string, ch chan interface{}) {
	c.wg.Add(1)
	defer c.wg.Done()

	pb := cony.NewPublisher(exchange, "")
	c.c.Publish(pb)

	send := func(msg interface{}) {
		if err := c.writeExchangeJsonMessage(pb, &KeyMessageJson{routing_key, msg}); err != nil {
			bmsg, _ := json.Marshal(msg)
			fmt.Printf("mqpublisher write failed msg: %v key: %v err: %v", bmsg, routing_key, err)
		}
	}

	// send message or wait close
	for c.run {
		select {
		case msg := <-ch:
			send(msg)
		case <-c.close:
		}
	}

	// send message immediate
	for len(ch) > 0 {
		send(<-ch)
	}

	// close channel
	close(ch)

	fmt.Printf("mqclient write end exchange: [%v]", exchange)
}

func (c *Client) ReadExchange(e cony.Exchange, f func(amqp.Delivery) error) {
	c.wg.Add(1)
	defer c.wg.Done()

	queue := &cony.Queue{AutoDelete: true}
	consumer := cony.NewConsumer(queue)
	deQueue := cony.DeclareQueue(queue)
	deBind := cony.DeclareBinding(cony.Binding{
		Queue:    queue,
		Exchange: e,
		Key:      "#",
	})

	c.c.Declare([]cony.Declaration{deQueue, deBind})
	c.c.Consume(consumer)

	for c.run {
		select {
		case msg := <-consumer.Deliveries():
			if err := f(msg); err != nil {
				b64 := base64.StdEncoding.EncodeToString(msg.Body)
				fmt.Printf("mqclient handler [%v] message: %v err: %v", queue.Name, b64, err)
			}

			msg.Ack(false)
		case err := <-consumer.Errors():
			if err == nil {
				continue
			}

			c.sendError(err)

			fmt.Printf("mqclient consumer [%v] err: %v", queue.Name, err)
		case <-c.close:
		}
	}

	fmt.Printf("mqclient read end queue: [%v]", queue.Name)
}

func (c *Client) sendError(err error) {
	if c.errcb != nil {
		c.errcb(err)
	}
}

func (c *Client) OnError(f func(error)) {
	c.errcb = f
}

func (c *Client) Close() {
	c.run = false
	close(c.close)
	c.wg.Wait()
	c.c.Close()
}
