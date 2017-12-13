package rest

const maxDataLen int = 512 * 1024 // 512K

type Config struct {
	BindAddr              string
	EnablePProf           bool
	EnableAPIMetrics      bool
	GraceShutdownTimeoutS int
}
