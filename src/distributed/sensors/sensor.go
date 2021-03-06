package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/limianwang/test-queue/src/distributed/dto"
	"github.com/limianwang/test-queue/src/distributed/qutils"
	"github.com/streadway/amqp"
)

var url = "amqp://guest:guest@0.0.0.0:5672"
var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "max")
var min = flag.Float64("min", 1., "min")
var stepSize = flag.Float64("step", 0.1, "max allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var value = r.Float64()*(*max-*min) + *min
var nom = (*max-*min)/2 + *min

func main() {
	flag.Parse()

	conn, ch := qutils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	dataQueue := qutils.GetQueue(*name, ch, false)
	// sensorQueue := qutils.GetQueue(qutils.SensorListQueue, ch)
	publishQueueName(ch)

	discoveryQueue := qutils.GetQueue("", ch, true)
	ch.QueueBind(
		discoveryQueue.Name,
		"",
		qutils.SensorDiscoveryExchange,
		false,
		nil,
	)

	go listenForDiscoveryRequests(discoveryQueue.Name, ch)

	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
	signal := time.Tick(dur)
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	for range signal {
		calcValue()
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}
		buf.Reset()
		enc.Encode(reading)

		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		ch.Publish(
			"",             // exchange string,
			dataQueue.Name, // key string,
			false,          // mandatory bool,
			false,          // immediate bool,
			msg,            // msg amqp.Publishing,
		)
		log.Printf("reading sent, Value %v \n", value)
	}
}

func publishQueueName(ch *amqp.Channel) {
	msg := amqp.Publishing{Body: []byte(*name)}

	ch.Publish(
		"amq.fanout", // exchange string,
		"",           // key string,
		false,        // mandatory bool,
		false,        // immediate bool,
		msg,          // msg amqp.Publishing,
	)

}

func listenForDiscoveryRequests(name string, ch *amqp.Channel) {
	msgs, _ := ch.Consume(
		name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	for range msgs {
		publishQueueName(ch)
	}
}

func calcValue() {
	var maxStep, minStep float64
	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += r.Float64()*(maxStep-minStep) + minStep
}
