/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.rdf.model.ModelFactory;
import com.hp.hpl.jena.rdf.model.RDFReader;
import com.hp.hpl.jena.util.FileManager;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import org.zeromq.ZMQ;

/**
 *
 * @author itiu
 */
public class Main
{

    private static Model get_message_from_file(String inputFileName) throws Exception
    {
        ByteArrayOutputStream baos = new ByteArrayOutputStream();
        // create an empty model
        Model message = ModelFactory.createDefaultModel();

        // use the FileManager to find the input file
        InputStream in = FileManager.get().open(inputFileName);
        if (in == null)
        {
            throw new IllegalArgumentException(
                    "File: " + inputFileName + " not found");
        }

        RDFReader r = message.getReader("N3");

        String baseURI = "";

        r.read(message, in, baseURI);

        message.read(in, null);

        in.close();

        return message;
    }

    public static void main(String[] args) throws Exception
    {
        ByteArrayOutputStream baos = new ByteArrayOutputStream();

        Model input_message = get_message_from_file("test001-in.n3");
        Model output_message = get_message_from_file("test001-out.n3");

        input_message.write(baos, "N3");

        String connectTo = "tcp://127.0.0.1:5555";
        ZMQ.Context ctx = ZMQ.context(1);
        ZMQ.Socket s = ctx.socket(ZMQ.REQ);

        s.connect(connectTo);

        long start = System.currentTimeMillis();

        byte data[] = baos.toByteArray();

        s.send(data, 0);
        String result = new String(s.recv(0));

        long end = System.currentTimeMillis();
        System.out.println("RES: (" + (end - start) + "[ms])\n" + result);

    }
}
