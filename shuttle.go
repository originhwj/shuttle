package main

import (
	logger "./utils/log"
	"bufio"
	"fmt"
	"net"
	"time"
	"flag"
	"os"
	"syscall"
	"net/http"
	"encoding/json"
	"io"
	"runtime"
	"regexp"
	"io/ioutil"
	"errors"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}

	Host      = ":8888"
	INBOX_LEN = 500
	env       = flag.String("env", "test", "dev environment")
	logPath   = flag.String("logPath", "./shuttle", "log path")
	port      = flag.String("http port", "12000", "http server port")
	log       *logger.Logger
	mux          map[string]*Router
	ErrorList = map[int64]string{
		-1: "读取配置失败",
		-2: "sig出错",

		-3: "查询服务器失败",
		-4: "操作失败",
	}
)

type Router struct {
	method     string
	handler    func(http.ResponseWriter, *http.Request, []string, []byte)
	sigCheck   bool
}

func init_log(log_path string) {
	filename := log_path + ".log"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Println("fail to create log file! err:", err)
		return
	}
	log = logger.New(file, "", logger.Ldate|logger.Ltime|logger.Lmicroseconds|logger.Lshortfile, logger.FINFO)
	syscall.Dup2(int(file.Fd()), 1)
	syscall.Dup2(int(file.Fd()), 2)
	logger.SetLogger(log)
}

func tcp_server() {
	var err error
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen err", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			return
		}
		terminal := Terminal{
			Conn:         conn,
			bw:           bufio.NewWriter(conn),
			br:           bufio.NewReader(conn),
			readTimeout:  10 * time.Second,
			writeTimeout: 10 * time.Second,
			inbox:        make(chan []byte, INBOX_LEN),
		}

		go terminal.Process()
		go terminal.write_loop()
	}
}

func time_sub(t time.Time) int64 {
	return int64(time.Now().Sub(t) / time.Millisecond)
}


func GetPostData(r *http.Request) ([]byte, error) {
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read post data error")
		return nil, errors.New("read post data error")
	}
	return post_data, nil
}

func response_err(err_code int64, w *http.ResponseWriter) {
	ret, _ := json.Marshal(map[string]interface{}{
		"err":     err_code,
		"err_msg": ErrorList[err_code],
	})
	(*w).Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(*w, string(ret))
}


type Auth2Server struct{}

func (*Auth2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start_time := time.Now()
	defer func() {
		if err := recover(); err != nil {
			log.Warn(time.Now(), "request:", r, time_sub(start_time), "500")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Internal server error!")
			log.Error(err)

			for i := 0; i < 10; i++ {
				funcName, file, line, ok := runtime.Caller(i)
				if ok {
					log.Printf("%v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
				}
			}

		}
	}()
	w.Header().Set("server", "shuttle_v0.1")
	for o_url, router := range mux {
		p_url := "^" + o_url + "$"
		re := regexp.MustCompile(p_url)
		m := re.FindStringSubmatch(r.URL.Path)
		if len(m) > 0 {
			//fmt.Println("m", m)
			if err := r.ParseForm(); err != nil {
				log.Error("err", err)
			}
			if r.Method == router.method {
				var post_data []byte
				if r.Method == "POST" {
					var err error
					post_data, err = GetPostData(r)
					if err != nil {
						response_err(-13, &w)
						return
					}
				}
				if router.sigCheck {

				}

				router.handler(w, r, m[1:], post_data)
				log.Warn(time.Now(), "request:", r, time_sub(start_time), "200")
			} else {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "Method not allowed!")
				log.Warn(time.Now(), "request:", r, time_sub(start_time), "400")
			}
			return
		}

	}
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "Page not found!")
	log.Warn(time.Now(), "request:", r, time_sub(start_time), "404")

}


func http_server(){

	server := http.Server{
		Addr:           ":" + *port,
		Handler:        &Auth2Server{},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	mux = make(map[string]*Router)
	//mux[`/`] = &Router{"GET", index_handler, false}
	//mux[`/(\d+)`] = &Router{"GET", test_handler, false}
	mux[`/test`] = &Router{"GET", test_handler, true}

	log.Warn("starting server...", server)
	err := server.ListenAndServe()
	raise_err(err)
}

func raise_err(err error) {
	if err != nil {
		log.Panic("panic err", err)
	}
}

func main() {

	flag.Parse()
	init_log(*logPath) //初始化日志

	go http_server()
	tcp_server()
}
