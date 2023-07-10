declare global {
    namespace NodeJS {
        interface ProcessEnv {
            KAFKA_BROKER_URL: string;
            KAFKA_TOPIC_NAME: string;
            KAFKA_GROUP_ID: string;
            KAFKA_SASL_USERNAME: string;
            KAFKA_SASL_PASSWORD: string;
        }
    }
}
export {};