package main

import (
	"amqp"
	"encoding/json"
	"fmt"
	"log"
	"magick"
	"os"
	"strings"
	// "time"
)

func failOnError(err error, msg string) {
	if err != nil {
		// fatal; printf Followed by a call to os.Exit(1)
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

	log.Printf("Trying to Connect")
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
		log.Printf("Cropping the image: %s", d.Body)
		image_message := []byte(d.Body)
		var image_attr map[string]string           // Used to hold the image attributes from Recd. JSON
		json.Unmarshal(image_message, &image_attr) // fetching Json.
		go process_msg(image_attr)
	}
}

func process_msg(image_attr map[string]string) {
	magick_image, err := magick.NewFromFile(image_attr["path"])
	geometry_string := image_attr["width"] + "x" + image_attr["width"] + "+" + image_attr["crop_x"] + "+" + image_attr["crop_y"]
	if err == nil {
		err = magick_image.Crop(geometry_string) //100x200+10+20
		if err == nil {
			file_path := strings.Replace(image_attr["path"], "original", geometry_string, 1)
			dest_path := file_path[0:strings.LastIndex(file_path, "/")]
			err = os.MkdirAll(dest_path+"/", 0777)
			magick_image.ToFile(file_path)
		} else {
			fmt.Println("ERRORS: %s", err)
		}
	} else {
		log.Printf("Image Magick File Error: %s", err)
	}
}

func main() {
	var (
		c   *Consumer
		err error
	)
	c, err = NewConsumer("amqp://guest:guest@localhost:5672/", "crop_router1", "direct", "image_cropper", "cropper", "c1")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer c.conn.Close()

}
