import zmq

def main(file_name):

    addr = 'tcp://172.17.4.64:5555'

    f = open (file_name, 'r')
    msg = f.read ()
    msg += '\0'

    c = zmq.Context()
    s = c.socket(zmq.REQ)
    s.connect(addr)

    print "Connecting to: ", addr        

    s.send(msg, copy=False)
    msg2 = s.recv(copy=False)

if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print "usage: prompt.py <addr>"
        raise SystemExit
    main(sys.argv[1])

    
