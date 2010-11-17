/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.rdf.model.ModelFactory;
import com.hp.hpl.jena.rdf.model.RDFReader;
import com.hp.hpl.jena.util.FileManager;
import java.io.BufferedWriter;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.FileWriter;
import java.io.IOException;
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
    public final static String prefix_rdf = "@prefix rdf:     <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .";
    public final static String prefix_rdfs = "@prefix rdfs:    <http://www.w3.org/2000/01/rdf-schema#> .";
    public final static String prefix_xsd = "@prefix xsd:     <http://www.w3.org/2001/XMLSchema#> .";
    public final static String prefix_msg = "@prefix msg:     <http://gost19.org/message#> .";
    public final static String prefix_auth = "@prefix auth:     <http://gost19.org/auth#> .";
    public final static String all_prefixs = prefix_rdf + "\n" + prefix_rdfs + "\n" + prefix_xsd + "\n" + prefix_msg + "\n" + prefix_auth + "\n";

    public static void main(String[] args) throws Exception
    {
        ByteArrayOutputStream baos = new ByteArrayOutputStream();

        Model input_message = get_message_from_file("test001-in.n3");
        Model output_message = get_message_from_file("test001-out.n3");
        Model result_message = null;

        input_message.write(baos, "N3");

        String connectTo = "tcp://127.0.0.1:5555";
        ZMQ.Context ctx = ZMQ.context(1);
        ZMQ.Socket s = ctx.socket(ZMQ.REQ);

        s.connect(connectTo);

        long start = System.currentTimeMillis();

        byte data[] = baos.toByteArray();

        String result = null;

        for (int i = 0; i < 1; i++)
        {
            s.send(data, 0);
            result = new String(s.recv(0));

            BufferedWriter out = new BufferedWriter(new FileWriter("test001-recieve.n3"));
            out.write(all_prefixs + result);
            out.close();


            result_message = get_message_from_file("test001-recieve.n3");
        }

        long end = System.currentTimeMillis();
        System.out.println("RES: (" + (end - start) + "[ms])\n" + result);

    }
}
