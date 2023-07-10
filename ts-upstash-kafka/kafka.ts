import {Kafka} from "kafkajs";

export const kafka = new Kafka({
    brokers: [process.env.KAFKA_BROKER_URL],
    sasl: {
        mechanism: 'scram-sha-256',
        username: process.env.KAFKA_SASL_USERNAME,
        password: process.env.KAFKA_SASL_PASSWORD,
    },
    ssl: true,
});

export const KAFKA_CONSUMER_GROUP = process.env.KAFKA_GROUP_ID ?? "";
export const KAFKA_ORDER_TOPIC = 'orders.events';

