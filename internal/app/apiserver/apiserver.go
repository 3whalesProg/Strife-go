package apiserver

type APIServer struct {
	config *Config
}

func New(config *Config) *APIServer {
	return &APIServer{
		config: config, // ЭТУ ПОЕБОТУ НЕ ТРОГЕАЕМ
	}
}

func (s *APIServer) Start() error {
	return nil
}
