package conf

type Config struct {
    Kafka   Kafka   `ini:"kafka"`
}

type Kafka struct {
    Dsn   string `ini:"dsn"`
    Topic string `ini:"topic"`
}
