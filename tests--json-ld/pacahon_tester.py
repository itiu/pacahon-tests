#!/usr/bin/python
# -*- coding: utf-8 -*-

import zmq, json_ld_processor as jlp


def main(name_test):

    my_context = {
		    "dc": "http://purl.org/dc/terms/",
		    "msg": "http://gost19.org/message#",
		    "rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		    "rdfs": "http://www.w3.org/2000/01/rdf-schema#",
		    "xsd": "http://www.w3.org/2001/XMLSchema#",
		    "auth": "http://www.gost19.org/auth#"
                 }


    addr = 'tcp://172.17.4.64:5555'

    f = open (name_test + "-in.json", 'r')
    msg_in = f.read ()
    f.close ()

    f = open (name_test + "-out.json", 'r')
    msg_out_eth = f.read ()
    f.close ()

    c = zmq.Context()
    s = c.socket(zmq.REQ)
    s.connect(addr)

    print "Connecting to: ", addr        

    s.send(msg_in + '\0', copy=False)
    msg_out_recv = s.recv(copy=False)
    
    f = open (name_test + "-recv.json", 'w')
    f.write (msg_out_recv.buffer)
    f.close ()

    p = jlp.Processor(my_context)       
    eth_triples = p.triples (msg_out_eth)
    recv_triples = p.triples (str(msg_out_recv.buffer))    

    # сравниваем эталонное сообщение и полученное 
    for t_et in eth_triples:
        print "ETH: <" + t_et["subj"] + "><" + t_et["prop"] + "><" + t_et["obj"] + ">"
        for t_recv in recv_triples:
	    print "RECV <" + t_recv["subj"] + "><" + t_recv["prop"] + "><" + t_recv["obj"] + ">"
	    ss_et = t_et["subj"]
	    fi = 0
	    if (ss_et.find ("http://gost19.org/message#") >= 0):
		fi = fi + 1
	    else:
		if ss_et == t_recv["subj"]:
		    fi = fi + 1
	    if fi > 0:		    
		if t_et["prop"] == t_recv["prop"]:
		    fi = fi + 1
		    if t_et["obj"] == t_recv["obj"]:
			fi = fi + 1
    	    print "FI:", fi
	
if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print "usage: pacahon_tester.py <testXXX>"
        raise SystemExit
    main(sys.argv[1])

    
