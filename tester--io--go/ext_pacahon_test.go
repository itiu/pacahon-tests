package main

import "fmt"
import zmq "github.com/alecthomas/gozmq"
import ioutil "io/ioutil"
import "strings"
import "os"
import "bufio"
import "json"
import "log"
import "time"
import "flag" // command line option parser
import uuid "github.com/dchest/uuid.go"

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

	var max_pull int = 10000
	var messages_in [10000]string
	var messages_out [10000]string
	var cur_in_pull int = 0
	//
	var msg_in_et string
	var msg_out_et string
	var prev_ct int64
	var count int
	var prev_count int

	var cc [50]chan *IOElement
	var cc_len int = 20

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
			cc_len = 1
		}

		// стартуем тестирующие нити
		for i := 0; i < cc_len; i++ {
			cc[i] = make(chan *IOElement)
			go ggg(cc[i], point, compare_result)
		}

		// выбираем данные для тестирующих нитей 

		for i := 0; i < len(ff); i++ {
			if strings.Contains(ff[i].Name, ".io") {

				log.Println("found io: ", ff[i].Name)

				f, err := os.Open(ff[i].Name)
				if err != nil {
					fmt.Printf("can't open file; err=%s\n", err.String())
					os.Exit(1)
				}

				defer f.Close()
				r, err := bufio.NewReaderSize(f, 4*1024)
				if err != nil {
					fmt.Println(err)
					return
				}

				var io_type bool = true

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
						} else {
							msg_out_et = line[pos:len(line)]
						}

						if io_type == false {

							ct := time.Nanoseconds()

							delta := ((float32(ct - prev_ct)) / 1e9)

							if ct-prev_ct > 1e9 {
								fmt.Println("count readed:", count, " cps:", (float32(count-prev_count))/delta)
								prev_ct = ct
								prev_count = count
							}

							messages_in[cur_in_pull] = msg_in_et
							messages_out[cur_in_pull] = msg_out_et
							cur_in_pull++
							if cur_in_pull >= max_pull {
								fmt.Println("piuuu, cur_in_pull=", cur_in_pull, " max_pull=", max_pull)

								var j int
								for i := 0; i < max_pull; i++ {
									var io_el IOElement
									io_el.in_msg = messages_in[i]
									io_el.out_msg = messages_out[i]

									cc[j] <- &io_el

									j++
									if j >= cc_len {
										j = 0
									}
								}

								cur_in_pull = 0
							}

							//							fmt.Println("msg_in_et ", msg_in_et)
							//							fmt.Println("msg_in_et ", msg_out_et)
							//							c <- &io_el

							//							fmt.Println("count ", count)

							count++

						}

					}
					line, err = r.ReadString('\r')
				}

				if err != os.EOF {
					fmt.Println(err)
					return
				}
			}
		}

	}

	//	time.Sleep(1000 * 60 * 1e9)
}

func ggg(c chan *IOElement, point string, compare_result string) {

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	sock_uuid := uuid.New()
	socket.SetSockOptString(zmq.IDENTITY, sock_uuid.String())
	socket.Connect(point)
	println("sock_uuid: ", sock_uuid.String())

	ticket := get_ticket(socket, "user", "9cXsvbvu8=")

	fmt.Println("ggg is waiting...")
	time.Sleep(10 * 1e9)

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

		m := f.(map[string]interface{})

		m["msg:ticket"] = ticket

		jsn_msg, err := json.Marshal(f)
		if err != nil {
			fmt.Println(err)
		}

		if jsn_msg != nil {

			msg_out_cmp, err := send_and_recieve(socket, jsn_msg, sock_uuid.String())

			if compare_result == "Y" {
				if msg_out_cmp == nil {
					fmt.Println(err)
				}

				var jsn_out_cmp []interface{}
				bb := make([]byte, len(msg_out_cmp)+2)
				bb[0] = '['
				copy(bb[1:len(msg_out_cmp)+1], msg_out_cmp)
				bb[len(msg_out_cmp)+1] = ']'

				err = json.Unmarshal(bb, &jsn_out_cmp)
				if err != nil {
					fmt.Println(err)
				}

				var jsn_out_et []interface{}
				err = json.Unmarshal([]byte(msg_out_et), &jsn_out_et)
				if err != nil {
					fmt.Println(err)
				}

				res, _ := cmp_msg_out("", jsn_out_et, "", jsn_out_cmp, 0)

				if res == false {
					fmt.Println("msg out_et != out ")

					fout_et, err := os.OpenFile("ou_et", os.O_WRONLY|os.O_CREATE, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_et.WriteString(msg_out_et)
					fout_et.Close()

					fout_cmp, err := os.OpenFile("ou_cmp", os.O_WRONLY|os.O_CREATE, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_cmp.Write(msg_out_cmp)
					fout_cmp.Close()

					fout_in, err := os.OpenFile("in_et", os.O_WRONLY|os.O_CREATE, 0666)
					if err != nil {
						fmt.Fprintf(os.Stderr, " %s\n", err)
						return
					}
					fout_in.WriteString(msg_in_et)
					fout_in.Close()

					os.Exit(1)
				}

			}
		}

	}

}

func cmp_msg_out(key_et string, msg_out_et interface{}, key_cmp string, msg_out_cmp interface{}, level int) (bool, int) {

	var is_level_down int = 0

	if msg_out_et == nil && msg_out_cmp != nil {
		return false, is_level_down
	}

	if msg_out_et != nil && msg_out_cmp == nil {
		return false, is_level_down
	}

	if msg_out_et == nil && msg_out_cmp == nil {
		return true, is_level_down
	}

	if key_et == "@" && level == 2 {
		return true, is_level_down
	}

	if key_et == "msg:reason" || key_et == "auth:ticket" {
		return true, is_level_down
	}

	switch vv := msg_out_et.(type) {
	case string:

		switch dd := msg_out_cmp.(type) {

		case string:
			if dd != vv {
				//				fmt.Println(key_et, ":", vv, " != ", dd, level)
				return false, is_level_down
			}
		default:
			//			fmt.Println(key_et, " => different type")
			return false, is_level_down
		}

	case int:
		dd := msg_out_cmp.(int)
		if dd != vv {
			//			fmt.Println(key_et, ":", vv, " != ", dd, level)
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
				return true, is_level_down
			}

			var is_local_level_down int = 0
			var res bool

			//			fmt.Println(level, " ***{", len(vv), " ", len(dd))
			for _, v := range vv {
				var is_found = false
				//				fmt.Println(level, "	cc=", v)
				for _, vr := range dd {
					//					fmt.Println(level, "		oo=", vr)
					res, is_local_level_down = cmp_msg_out("", v, "", vr, level+1)
					is_level_down = 1
					if res == true {
						is_found = true
						break
					}

				}

				if is_found == false {
					if is_local_level_down == 1 && level > 0 {
						fmt.Println(level, " ", key_et, " : et != cmp, ", v)
						fmt.Println(level, " ", key_et, " => len (res):", len(dd), " != len (et):", len(vv))
					}
					return false, is_local_level_down
				}
			}
			//			fmt.Println(level, " }***")

			//			if count_found > 1 {
			//				fmt.Println(count_found, ":", count_compared)
			//			}

			if len(dd) != len(vv) {
				fmt.Println(level, " ", key_et, " => len (res):", len(dd), " != len (et):", len(vv))
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
			//			fmt.Println(key_et, " => different type")
			return false, is_level_down

		}

	case map[string]interface{}:
		for k, v := range vv {
			switch dd := msg_out_cmp.(type) {
			case map[string]interface{}:

				//			dd := msg_out_cmp.(map[string]interface{})
				res, _ := cmp_msg_out(k, v, k, dd[k], level+1)
				is_level_down = 1
				if res == false {
					return false, is_level_down
				}
			default:
				//				fmt.Println(key_et, " => different type")
				return false, is_level_down
			}
		}

	default:
		fmt.Println(msg_out_et, "is of a type I don't know how to handle, ", vv)

	}

	return true, is_level_down
}

func get_ticket(socket zmq.Socket, login string, credential string) string {
	msg_uuid := uuid.New()

	var msg string
	msg = "{\n\"@\" : \"msg:M" + msg_uuid.String() + "\", \n\"a\" : \"msg:Message\",\n\"msg:sender\" : \"ext_pacahon_test\",\n\"msg:reciever\" : \"pacahon\",\n" + "\"msg:command\" : \"get_ticket\",\n" + "\"msg:args\" :\n" + "{\n" + "  \"auth:login\" : \"" + login + "\",\n  \"auth:credential\" : \"" + credential + "\"\n}\n}"

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

func send_and_recieve(socket zmq.Socket, in_msg []byte, id string) (res []byte, err os.Error) {
	//		println("in_msg: ", string (in_msg))
	//	println("send ", id)
	var repeat bool
	var r0 []byte
	var err0 os.Error

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
