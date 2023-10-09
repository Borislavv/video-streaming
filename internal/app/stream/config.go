package stream

type config struct {
	// >>> SERVER <<<
	// Host is an WebSocket server serving host.
	Host string `env:"STREAMING_SERVER_HOST" envDefault:"0.0.0.0"`
	// Port is an WebSocket server serving port.
	Port string `env:"STREAMING_SERVER_PORT" envDefault:"9988"`
	// Transport is an WebSocket server transport protocol.
	// If you are not concerned about the loss part of packets and this is not a problem for you, then use the UDP,
	// because this will give you a performance gain (due to the server will not check of packages number and them ordering).
	// Otherwise, if your data needs to be in safe, and you cannot afford to lose it, use the TCP.
	Transport string `env:"STREAMING_SERVER_TRANSPORT_PROTOCOL" envDefault:"tcp"`
	// database
	MongoUri string `env:"MONGO_URI" envDefault:"mongodb://mongodb:27017/streaming"`
	MongoDb  string `env:"MONGO_DATABASE" envDefault:"streaming"`
	// file reader
	ChunkSize int `env:"FILE_READER_CHUNK_SIZE" envDefault:"1048576"` // by default: chunk size is 1mb.
}
