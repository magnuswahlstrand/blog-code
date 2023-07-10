import {kafka, KAFKA_CONSUMER_GROUP, KAFKA_ORDER_TOPIC} from './kafka'

const TIMEOUT= 10;

const handleEachMessage = async ({topic, partition, message}: any) => {
    if (!message.value) {
        throw new Error('No message value')
    }

    console.log({
        partition,
        offset: message.offset,
        value: JSON.parse(message.value.toString()),
    })
}

(async () => {
    let timer: NodeJS.Timeout | null = null;

    const consumer = kafka.consumer({
        groupId: KAFKA_CONSUMER_GROUP,
    })

    await consumer.connect()
    await consumer.subscribe({
        topic: KAFKA_ORDER_TOPIC,
        fromBeginning: true
    })

    const resetTimer = () => {
        if (timer) {
            clearTimeout(timer); // Reset the timer
        }
        timer = setTimeout(async () => {
            console.log(`No new messages within ${TIMEOUT} seconds. Shutting down...`);
            await consumer.disconnect();
        }, TIMEOUT * 1000);
    }

    resetTimer();
    console.log('Consuming messages...')
    await consumer.run({
        eachMessage: handleEachMessage,
    })
    console.log('Shutting down')

})()


console.log('Here')