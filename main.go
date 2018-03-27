package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	go client()
	go server()

	var a string
	fmt.Scanln(&a)
}

func client() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name, //queue string,
		"",     // consumer string,
		true,   // autoAck bool,
		false,  // exclusive bool,
		false,  // noLocal bool,
		false,  // noWait bool,
		nil,    // args amqp.Table,
	)
	failOnError(err, "failed to register a consumer")

	for msg := range msgs {
		log.Printf("Reveived message %s", msg.Body)
	}
}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello RabbitMQ"),
	}

	for {
		ch.Publish("", q.Name, false, false, msg)
	}
}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@0.0.0.0:5672")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"hello", // name string
		false,   // durable bool
		false,   // autoDelete pool,
		false,   // exclusive bool,
		false,   // noWait bool,
		nil,     // args amqp.Table,
	)
	failOnError(err, "Failed to declare a queue")

	return conn, ch, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
