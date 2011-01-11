#!/usr/bin/python

import zmq

def main(name_test):

    addr = 'tcp://172.17.4.64:5555'

    f = open (name_test + "-in.json", 'r')
    msg = f.read ()
    msg += '\0'
    f.close ()

    c = zmq.Context()
    s = c.socket(zmq.REQ)
    s.connect(addr)

    print "Connecting to: ", addr        

    s.send(msg, copy=False)
    msg2 = s.recv(copy=False)
    
    f = open (name_test + "-recv.json", 'w')
    f.write (msg2.buffer)
    f.close ()
    

if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print "usage: pacahon_tester.py <testXXX>"
        raise SystemExit
    main(sys.argv[1])

    
