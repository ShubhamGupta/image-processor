package main

import (
	"amqp"
	"fmt"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
}

func NewConsumer(url, exchange, exchange_type, queueName, key, ctag string) (*Consumer, error) {
	var queue amqp.Queue
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
	}
	var err error

	fmt.Println("Trying to Connect")
	c.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}
	log.Printf("Got the connection, getting the channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}
	log.Printf("Got the Channel, declaring the exchange: %s", exchange)
	if err = c.channel.ExchangeDeclare(exchange, exchange_type, true, true, false, false, nil); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}
	log.Printf("declared Exchange, declaring Queue %q", queueName)
	if queue, err = c.channel.QueueDeclare(queueName, true, true, false, false, nil); err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)
	if err = c.channel.QueueBind(queue.Name, key, exchange, false, nil); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(queue.Name, c.tag, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	handle(deliveries)

	return c, nil

}

func handle(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		fmt.Println("Sending for Processing...")
		go process_msg(d)
		fmt.Println(time.Now())
	}
	fmt.Println("Closing the connection")
}

func process_msg(d amqp.Delivery) {
	// fmt.Println(d.MessageCount)
	fmt.Printf("Content: %s", d.Body)
}

func main() {
	var (
		c   *Consumer
		err error
	)
	c, err = NewConsumer("amqp://guest:guest@localhost:5672/", "crop_router1", "direct", "image_cropper1", "cropper", "c1")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer c.conn.Close()

}
