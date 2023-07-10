import {Kafka} from "kafkajs";

const kafka = new Kafka({
    brokers: ['humorous-seahorse-8248-eu1-kafka.upstash.io:9092'],
    sasl: {
        mechanism: 'scram-sha-256',
        username: 'aHVtb3JvdXMtc2VhaG9yc2UtODI0OCQTvJGUBlmIsXYkLHZjNpjKZXVDw62_ie8',
        password: 'cb6026c0d6d644a1b006d407be62f409',
    },
    ssl: true,
});

(async () => {

    const producer = kafka.producer();

    await producer.connect()

    await producer.send({
        topic: "orders.events",
        messages: [
            {
                value: JSON.stringify({
                    id: '1',
                    name: 'Order 1',
                    amount: 100,

                })
            },
        ],
    })
    console.log('Published successfully')
    await producer.disconnect()
})()


