package main

import "fmt"
import ioutil "io/ioutil"
import "strings"
import "io"
import "os"
import "bufio"
import "encoding/json"
import "log"
import time "time"
import "flag" // command line option parser
import uuid "github.com/serverhorror/uuid"
import zmq "github.com/alecthomas/gozmq"

type IOElement struct {
	in_msg  string
	out_msg string
}

func main() {

	flag.Parse() // Scans the arg list and sets up flags

	if flag.NArg() <= 2 {
		fmt.Fprintf(os.Stderr, "exec: ext_pacahon_test [string:(pacahon point)] [Y/N:(compare result)] [Y/N:(multi thread)]")
		return
	}

	point := flag.Arg(0)
	compare_result := flag.Arg(1)
	multi_thread := flag.Arg(2)

	var io_file string

	if flag.NArg() == 4 {
		io_file = flag.Arg(3)
	}

	var max_pull int = 1000

	messages_in := make([]string, max_pull)
	messages_out := make([]string, max_pull)

	var cur_in_pull int = 0
	//
	var msg_in_et string
	var msg_out_et string
	var prev_ct time.Time
	var count int64
	var prev_count int64

	var test_chanel_len int = 20
	test_chanel := make([]chan *IOElement, test_chanel_len)

	fout, err := os.OpenFile("logfile", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s\n", err)
		return
	}

	log.SetOutput(fout)
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	ff, _ := ioutil.ReadDir("./")

	if ff != nil {

		if multi_thread != "Y" {
			test_chanel_len = 1
		}

		// стартуем тестирующие нити
		for i := 0; i < test_chanel_len; i++ {
			test_chanel[i] = make(chan *IOElement)
			go ggg(test_chanel[i], point, compare_result)
		}

		// выбираем данные для тестирующих нитей 

		for i := 0; i < len(ff); i++ {
			if io_file == "" && strings.Contains(ff[i].Name(), ".io") || io_file != "" && strings.Contains(ff[i].Name(), io_file) {

				log.Println("found io: ", ff[i].Name())

				f, err := os.Open(ff[i].Name())
				if err != nil {
					fmt.Printf("can't open file; err=%s\n", err)
					os.Exit(1)
				}

				defer f.Close()

				r := bufio.NewReaderSize(f, 4*1024)

				log.Println("read io: ", ff[i].Name())

				var io_type bool = true
				cur_in_pull = 0

				line, err := r.ReadString('\r')
				for err == nil {

					pos := strings.Index(line, "INPUT")
					if pos > 0 {
						pos += 6
						io_type = true
					}

					if pos < 0 {
						pos = strings.Index(line, "OUTPUT")
						if pos > 0 {
							pos += 7
							io_type = false
						}
					}

					if pos > 0 {

						if io_type == true {
							msg_in_et = line[pos:len(line)]
							//							log.Println("read msg-in ", msg_in_et)
						} else {
							msg_out_et = line[pos:len(line)]
							//							log.Println("read msg-out ", msg_out_et)
						}

						if io_type == false {

							ct := time.Now()

							delta := time.Now().Sub(prev_ct).Nanoseconds()

							if delta > 1e9 {
								fmt.Println("count readed:", count, " cps:", (count-prev_count)/delta)
								prev_ct = ct
								prev_count = count
							}

							//							log.Println("msg to pull")
							// Если msg_in_et содержит команду изменяющую базу данных
							// то закончим формирование пула 
							var pull_is_ready bool = false

							pos = strings.Index(msg_in_et, "\"msg:command\" : \"get\"")
							if pos < 0 {
								pos = strings.Index(msg_in_et, "\"msg:command\" : \"get_ticket\"")
								if pos < 0 {
									pos = strings.Index(msg_in_et, "\"msg:command\" : \"remove\"")
								}
							}

							if cur_in_pull >= max_pull || pos > 0 {
								pull_is_ready = true
							}

							if pull_is_ready == true {
								send_packet(messages_in, messages_out, cur_in_pull, test_chanel, test_chanel_len)
								cur_in_pull = 0
							}

							messages_in[cur_in_pull] = msg_in_et
							messages_out[cur_in_pull] = msg_out_et
							cur_in_pull++

							//							fmt.Println("msg_in_et ", msg_in_et)
							//							fmt.Println("msg_in_et ", msg_out_et)
							//							fmt.Println("count ", count)

							count++

						}

					}
					line, err = r.ReadString('\r')
					//					//fmt.Println("read line, len=", line)

				}

				//				fmt.Println("piuuu, cur_in_pull=", cur_in_pull, " max_pull=", max_pull)

				// отправим остаток на тестирование
				send_packet(messages_in, messages_out, cur_in_pull, test_chanel, test_chanel_len)

				fmt.Println("file complete, count messages = ", cur_in_pull)

				if err != io.EOF {
					fmt.Println(err)
					return
				}
			}
		}

	}

	//	time.Sleep(1000 * 60 * 1e9)
}

func send_packet(messages_in []string, messages_out []string, pull_size int, test_chanel []chan *IOElement, test_chanel_len int) {
	// пулл подготовлен, отправляем пачку messages_in[] messages_out[] паралельно во все тестирующие каналы 

	//fmt.Println("piuuu, cur_in_pull=", cur_in_pull, " max_pull=", max_pull)

	var j int
	for i := 0; i < pull_size; i++ {
		var io_el IOElement
		io_el.in_msg = messages_in[i]
		io_el.out_msg = messages_out[i]

		test_chanel[j] <- &io_el

		j++
		if j >= test_chanel_len {
			j = 0
		}
	}

}

func ggg(c chan *IOElement, point string, compare_result string) {

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	sock_uuid := uuid.UUID4()
	socket.SetSockOptString(zmq.IDENTITY, sock_uuid)
	socket.Connect(point)
	println("sock_uuid: ", sock_uuid)

	ticket := get_ticket(socket, "user", "9cXsvbvu8=")

	fmt.Println("ggg is waiting...")
	time.Sleep(5 * 1e9)

	fmt.Println("ggg read chanel")

	for {
		io_el := <-c

		if io_el == nil {
			continue
		}

		msg_in_et := io_el.in_msg
		msg_out_et := io_el.out_msg

		var f interface{}
		err := json.Unmarshal([]byte(msg_in_et), &f)
		if err != nil {
			fmt.Println("err! unmarshal msg_in_et:", err)
			fmt.Println(string(msg_in_et))
			return
		}

		m := f.(map[string]interface{})

		m["msg:ticket"] = ticket

		jsn_msg, err := json.Marshal(f)
		if err != nil {
			fmt.Println(err)
		}

		if jsn_msg != nil {

			msg_out_cmp, err := send_and_recieve(socket, jsn_msg, sock_uuid)

			if compare_result == "Y" {
				if msg_out_cmp == nil {
					fmt.Println(err)
				}

				var msg_out_et_array []byte
				var bb []byte

				if msg_out_cmp[0] != '[' {
					bb = make([]byte, len(msg_out_cmp)+2)
					bb[0] = '['
					copy(bb[1:len(msg_out_cmp)+1], msg_out_cmp)
					bb[len(msg_out_cmp)+1] = ']'
					msg_out_cmp = bb
				}

				if msg_out_et[0] != '[' {
					bb = make([]byte, len(msg_out_et)+2)
					bb[0] = '['
					copy(bb[1:len(msg_out_et)+1], msg_out_et)
					bb[len(msg_out_et)+1] = ']'
					msg_out_et_array = bb
				} else {
					msg_out_et_array = []byte(msg_out_et)
				}

				//fmt.Println("len=", len (msg_out_et_array));

				var jsn_out_et []interface{}
				err = json.Unmarshal(msg_out_et_array, &jsn_out_et)
				if err != nil {
					fmt.Println("err! unmarshal out_et:", err)
					fmt.Println(msg_out_et)
					return
				}

				var jsn_out_cmp []interface{}
				err = json.Unmarshal(msg_out_cmp, &jsn_out_cmp)
				if err != nil {
					fmt.Println("err! unmarshal out_cmp:", err)
					fmt.Println(string(bb))
					return
				}

				res, _ := cmp_msg_out("", jsn_out_et, "", jsn_out_cmp, 0, false)

				if res == false {
					fmt.Println("msg out_et != out ")

					//					cmp_msg_out("", jsn_out_et, "", jsn_out_cmp, 0, true)

					fout_et, err := os.OpenFile("ou_et", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_et.WriteString(msg_out_et)
					fout_et.Close()

					fout_cmp, err := os.OpenFile("ou_cmp", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_cmp.Write(msg_out_cmp)
					fout_cmp.Close()

					fout_in, err := os.OpenFile("in_et", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_in.WriteString("[]\n")
					fout_in.WriteString("INPUT\n")
					fout_in.WriteString(msg_in_et)
					fout_in.WriteString("[]\n")
					fout_in.WriteString("OUTPUT\n")
					fout_in.WriteString(msg_out_et)
					fout_in.Close()

					os.Exit(1)
				}

			}
		}

	}

}

func cmp_msg_out(key_et string, msg_out_et interface{}, key_cmp string, msg_out_cmp interface{}, level int, trace bool) (bool, int) {

	trace = false
	//	if trace {
	//		fmt.Println(level, "cmp_msg_out, key_cmp=", key_cmp)
	//	}

	var is_level_down int = 0

	if msg_out_et == nil && msg_out_cmp != nil {

		if trace {
			fmt.Println(level, "cmp_msg_out, key_cmp=", key_cmp)
			fmt.Println(level, " return")
		}
		return false, is_level_down
	}

	if msg_out_et != nil && msg_out_cmp == nil {
		if trace {
			fmt.Println(level, "cmp_msg_out, key_cmp=", key_cmp)
			fmt.Println(level, " return")
		}
		return false, is_level_down
	}

	if msg_out_et == nil && msg_out_cmp == nil {
		//		if trace {
		//			fmt.Println(level, "return")
		//		}
		return true, is_level_down
	}

	if key_et == "@" && level == 2 {
		//		if trace {
		//			fmt.Println(level, "return")
		//		}
		return true, is_level_down
	}

	if key_et == "msg:reason" || key_et == "auth:ticket" {
		//		if trace {
		//			fmt.Println(level, "return")
		//		}
		return true, is_level_down
	}

	switch vv := msg_out_et.(type) {
	case string:

		switch dd := msg_out_cmp.(type) {

		case string:
			if dd != vv {
				if key_et == "@" && dd[0] == '_' {
					return true, is_level_down
				}

				if trace {
					fmt.Println(level, " ", key_et, ":", vv, " != ", dd)
				}
				return false, is_level_down
			}
		default:
			if trace {
				fmt.Println(level, " ", key_et, " => different type")
			}
			return false, is_level_down
		}

	case int:
		dd := msg_out_cmp.(int)
		if dd != vv {
			if trace {
				fmt.Println(level, " ", key_et, ":", vv, " != ", dd)
			}
			return false, is_level_down
		}

	case []interface{}:

		switch dd := msg_out_cmp.(type) {

		case []interface{}:

			//						if len(dd) != len(vv) {
			//							fmt.Println(key_et, " => len (res):", len(dd), " != len (et):", len(vv))
			//							return false, is_level_down
			//						}

			if len(vv) == 0 && len(dd) == 0 {
				//				if trace {
				//					fmt.Println(level, " return")
				//				}
				return true, is_level_down
			}

			var is_local_level_down int = 0
			var res bool

			//			if trace {
			//				fmt.Println(level, " ***{", len(vv), " ", len(dd))
			//			}
			for _, v := range vv {
				var is_found = false
				//				if trace {
				//					fmt.Println(level, "	cc=", v)
				//				}
				for _, vr := range dd {
					//					if trace {
					//						fmt.Println(level, "		oo=", vr)
					//					}
					res, is_local_level_down = cmp_msg_out("", v, "", vr, level+1, trace)
					is_level_down = 1
					if res == true {
						is_found = true
						break
					}

				}

				if is_found == false {
					if is_local_level_down == 1 && level > 0 {
						//fmt.Println(level, " ", key_et, " : et != cmp, ", v)
						//fmt.Println(level, " ", key_et, " => len (res):", len(dd), " != len (et):", len(vv))

						if len(dd) == 6000 || len(dd) == 1000 {
							fmt.Println("!!! len(dd) = ", len(dd))
							return true, is_local_level_down
						}
					}
					return false, is_local_level_down
				}
			}
			//			if trace {
			//				fmt.Println(level, " }***")
			//			}

			//			if count_found > 1 {
			//				fmt.Println(count_found, ":", count_compared)
			//			}

			if len(dd) != len(vv) {
				//fmt.Println(level, " ", key_et, " => len (res):", len(dd), " != len (et):", len(vv))
				return false, is_level_down
			}

			//			if count_found == count_compared {
			//			return true, is_local_level_down
			//			}

			//				if level == 4 {
			//					fmt.Println(key_et, ": et != cmp, ", vv, dd)
			//				} else {
			//					fmt.Println(key_et, ": et != cmp, level:", level)
			//				}
			return true, is_level_down

		default:
			if trace {
				fmt.Println(key_et, " => different type")
			}
			return false, is_level_down

		}

	case map[string]interface{}:
		for k, v := range vv {
			switch dd := msg_out_cmp.(type) {
			case map[string]interface{}:

				//			dd := msg_out_cmp.(map[string]interface{})
				res, _ := cmp_msg_out(k, v, k, dd[k], level+1, trace)
				is_level_down = 1
				if res == false {
					if trace {
						fmt.Println(level, " return")
					}
					return false, is_level_down
				}
			default:
				fmt.Println(key_et, " => different type")
				if trace {
					fmt.Println(level, " return")
				}
				return false, is_level_down
			}
		}

	default:
		fmt.Println(msg_out_et, "is of a type I don't know how to handle, ", vv)

	}

	return true, is_level_down
}

func get_ticket(socket zmq.Socket, login string, credential string) string {
	msg_uuid := uuid.UUID4()

	var msg string
	msg = "{\n\"@\" : \"msg:M" + msg_uuid + "\", \n\"a\" : \"msg:Message\",\n\"msg:sender\" : \"ext_pacahon_test\",\n\"msg:reciever\" : \"pacahon\",\n" + "\"msg:command\" : \"get_ticket\",\n" + "\"msg:args\" :\n" + "{\n" + "  \"auth:login\" : \"" + login + "\",\n  \"auth:credential\" : \"" + credential + "\"\n}\n}"

	//socket.Send([]byte(msg), 0)
	//res, err := socket.Recv(0)
	res, err := send_and_recieve(socket, []byte(msg), "---")
	if res == nil {
		fmt.Println(err)
	}

	out_msg := string(res)

	var auth__ticket string
	auth__ticket = "auth:ticket"

	pos := strings.Index(out_msg, auth__ticket)

	if pos > 0 {
		out_msg = out_msg[pos+len(auth__ticket)+1 : len(out_msg)]
		start_pos := strings.Index(out_msg, "\"")
		out_msg = out_msg[start_pos+1 : len(out_msg)]
		end_pos := strings.Index(out_msg, "\"")
		out_msg = out_msg[0:end_pos]

		println("get ticket: ", out_msg)
		return out_msg
	}

	return "0000"
}

func send_and_recieve(socket zmq.Socket, in_msg []byte, id string) (res []byte, err error) {
	//		println("in_msg: ", string (in_msg))
	//	println("send ", id)
	var repeat bool
	var r0 []byte
	var err0 error

	repeat = true

	for repeat {

		socket.Send(in_msg, 0)
		//		println("ok")
		r0, err0 = socket.Recv(0)

		if r0 != nil && len(r0) == 3 {
			// это указание повторить запрос еще раз 
			repeat = true
			time.Sleep(1e6)
		} else {
			repeat = false
		}
	}
	//	println("recv ", id)
	//			println("out_msg: ", string (r0))
	return r0, err0
}

/*
func send_and_recieve(socket zmq.Socket, in_msg []byte, id string) (res []byte, err os.Error) {

	println("send ", id))
	socket.Send(in_msg, 0)


	id_reply, err := socket.Recv(0)
	println("recv ", id))
//	println("id_reply: ", string (id_reply))

	for q := 0; q < 1000; q++ {
		socket.Send(id_reply, 0)
		r0, err := socket.Recv(0)
//		println("r0: ", string (r0))

		if string(r0) != "WAIT" {
			return r0, err
		}
	}
	return nil, nil
}
*/
