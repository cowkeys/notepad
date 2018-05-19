package mq

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/assembla/cony"
	"github.com/micro/protobuf/proto"
)

type ClientConOpt func(*ClientCon)
type ErrorHandler func(error)

func WithErrorHandler(fn ErrorHandler) ClientConOpt {
	return func(con *ClientCon) {
		con.ecb = fn
	}
}

func WithOptions(opts ...cony.ClientOpt) ClientConOpt {
	return func(con *ClientCon) {
		con.opts = append(con.opts, opts...)
	}
}

type ConSubscribeOpt func(*ConSubscribe)
type ConSubscribeHandler func(*Delivery) error

func WithAutoAck(ack bool) ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.ack = ack
	}
}

func WithQueue(q *cony.Queue) ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.q = q
	}
}

func WithBinding(b *cony.Binding) ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.b = b
	}
}

func WithQueueName(name string) ConSubscribeOpt {
	q := &cony.Queue{Name: name}
	return WithQueue(q)
}

func WithBindingKey(q *cony.Queue, e *cony.Exchange, key string) ConSubscribeOpt {
	return WithBinding(&cony.Binding{
		Exchange: *e,
		Queue:    q,
		Key:      key,
	})
}

func WithDeclareBinding() ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.c.con.Declare([]cony.Declaration{
			cony.DeclareBinding(*s.b),
		})
	}
}

func WithDeclareQueueExchangeBinding() ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.c.con.Declare([]cony.Declaration{
			cony.DeclareQueue(s.b.Queue),
			cony.DeclareExchange(s.b.Exchange),
			cony.DeclareBinding(*s.b),
		})
	}
}

func WithDeclareQueue() ConSubscribeOpt {
	return func(s *ConSubscribe) {
		s.c.con.Declare([]cony.Declaration{
			cony.DeclareQueue(s.q),
		})
	}
}

func WithTemporaryQueue(exchangeName, key string) ConSubscribeOpt {
	e := &cony.Exchange{Name: exchangeName}
	q := &cony.Queue{AutoDelete: true}
	return func(s *ConSubscribe) {
		WithQueue(q)(s)
		WithBindingKey(q, e, key)(s)
		WithDeclareQueue()(s)
		WithDeclareBinding()(s)
	}
}

func WithTemporaryQueueAnyKey(exchangeName string) ConSubscribeOpt {
	return WithTemporaryQueue(exchangeName, "#")
}

func WithProtoUnmarshal(val interface{}) ConSubscribeOpt {
	typ := reflect.TypeOf(val)
	return func(s *ConSubscribe) {
		f := func(d *Delivery) error {
			n := reflect.New(typ)
			p := n.Interface().(proto.Message)

			if err := proto.Unmarshal(d.Delivery.Body, p); err != nil {
				return fmt.Errorf("proto unmarshal err: %v", err)
			}

			d.Proto = p
			return nil
		}

		s.cb = append(s.cb, f)
	}
}

func WithProtoHandler(f interface{}) ConSubscribeOpt {
	return func(s *ConSubscribe) {
		//fw := fmpb.WrapProtoUnmarshal(f)
		fn := func(d *Delivery) error { return fw(d.Body) }
		WithHandler(fn)(s)
	}
}

func WithHandler(handler ConSubscribeHandler) ConSubscribeOpt {
	return func(s *ConSubscribe) {
		f := func(d *Delivery) error {
			var gerr error

			func() {
				defer func() {
					if err := recover(); err != nil {
						gerr = fmt.Errorf("panic during handler err: %v", err)
					}
				}()

				gerr = handler(d)
				if gerr != nil {
					b64 := base64.StdEncoding.EncodeToString(d.Delivery.Body)
					gerr = fmt.Errorf("handler message: %v err: %v", b64, gerr)
				}
			}()

			if gerr == nil && s.ack {
				if err := d.Ack(); err != nil {
					gerr = fmt.Errorf("ack err: %v", err)
				}
			}

			return gerr
		}

		s.cb = append(s.cb, f)
	}
}

type ConPublisherOpt func(*ConPublisher)
type ConPublisherFilter func(*Publishing) error

func WithCachedSize(size int) ConPublisherOpt {
	return func(p *ConPublisher) {
		p.q = make(chan *Publishing, size)
	}
}

func WithExchange(e *cony.Exchange) ConPublisherOpt {
	return func(p *ConPublisher) {
		p.e = e
	}
}

func WithExchangeName(name string) ConPublisherOpt {
	return WithExchange(&cony.Exchange{
		Name: name,
	})
}

func WithExchangeNameKind(name, kind string) ConPublisherOpt {
	return WithExchange(&cony.Exchange{
		Name: name,
		Kind: kind,
	})
}

func WithExchangeNameKindDurable(name, kind string, durable bool) ConPublisherOpt {
	return WithExchange(&cony.Exchange{
		Name:    name,
		Kind:    kind,
		Durable: durable,
	})
}

func WithDeclareExchange() ConPublisherOpt {
	return func(p *ConPublisher) {
		p.c.con.Declare([]cony.Declaration{
			cony.DeclareExchange(*p.e),
		})
	}
}

func WithProtoMarshal() ConPublisherOpt {
	return func(p *ConPublisher) {
		f := func(p *Publishing) error {
			if p.Proto == nil {
				return nil
			}

			d, err := proto.Marshal(p.Proto)
			if err != nil {
				return fmt.Errorf("proto marshal err: %v", err)
			}

			p.Body = d
			return nil
		}

		p.f = append(p.f, f)
	}
}
