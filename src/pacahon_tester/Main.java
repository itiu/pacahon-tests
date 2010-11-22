/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.datatypes.RDFDatatype;
import com.hp.hpl.jena.graph.Graph;
import com.hp.hpl.jena.graph.Node;
import com.hp.hpl.jena.graph.Triple;
import com.hp.hpl.jena.graph.impl.LiteralLabel;
import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.rdf.model.ModelFactory;
import com.hp.hpl.jena.rdf.model.RDFReader;
import com.hp.hpl.jena.util.FileManager;
import com.hp.hpl.jena.util.iterator.ExtendedIterator;
import java.io.BufferedWriter;
import java.io.ByteArrayOutputStream;
import java.io.FileWriter;
import java.io.InputStream;
import java.util.HashMap;
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
    public final static String ns_full_rdf = "http://www.w3.org/1999/02/22-rdf-syntax-ns#";
    public final static String ns_full_msg = "http://gost19.org/message#";
    public final static String prefix_rdf = "@prefix rdf:     <" + ns_full_rdf + "> .";
    public final static String prefix_rdfs = "@prefix rdfs:    <http://www.w3.org/2000/01/rdf-schema#> .";
    public final static String prefix_xsd = "@prefix xsd:     <http://www.w3.org/2001/XMLSchema#> .";
    public final static String prefix_msg = "@prefix msg:     <" + ns_full_msg + "> .";
    public final static String prefix_auth = "@prefix auth:     <http://gost19.org/auth#> .";
    public final static String all_prefixs = prefix_rdf + "\n" + prefix_rdfs + "\n" + prefix_xsd + "\n" + prefix_msg + "\n" + prefix_auth + "\n";
    private static HashMap<String, String> nsShort__nsFull;

    public static void main(String[] args) throws Exception
    {
        nsShort__nsFull = new HashMap<String, String>();

        nsShort__nsFull.put("rdf", "http://www.w3.org/1999/02/22-rdf-syntax-ns#");
        nsShort__nsFull.put("msg", "http://gost19.org/message#");

        String test_name = "test001";

        if (args.length > 0)
        {
            test_name = args[0];
        }

        ByteArrayOutputStream baos = new ByteArrayOutputStream();

        Model m_input_message = get_message_from_file(test_name + "-in.n3");
        Graph input_message = m_input_message.getGraph();
        Graph output_message = get_message_from_file(test_name + "-out.n3").getGraph();

        Graph result_message = null;

        m_input_message.write(baos, "N3");

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

            BufferedWriter out = new BufferedWriter(new FileWriter(test_name + "-recieve.n3"));
            out.write(all_prefixs + "\n" + result);
            out.close();


            // сравниваем результаты полученные и результаты из шаблона xxxx-out.n3
            result_message = get_message_from_file(test_name + "-recieve.n3").getGraph();
            if (cmp_results(input_message, result_message, null, output_message, null) == false)
            {
                throw new Exception("result != ethalon");
            } else
            {
            }


//            ExtendedIterator<Triple> it = graph.find(null, Node.createURI(ns_msg + "reciever"), null);

//            while (it.hasNext())
//            {
//                Triple tt = it.next();
//                System.out.println(tt.toString());
//            }
        }

        long end = System.currentTimeMillis();
        System.out.println("RES: (" + (end - start) + "[ms])\n" + result);

    }

    public static boolean cmp_results(Graph input, Graph result, Node result_subject, Graph ethalon, Node ethalon_subject) throws Exception
    {
        if (ethalon_subject == null)
        {
            ExtendedIterator<Triple> it = ethalon.find(null, Node.createURI(ns_full_rdf + "type"), Node.createURI(ns_full_msg + "Message"));

            if (it.hasNext())
            {
                Triple tt = it.next();
                ethalon_subject = tt.getSubject();
                System.out.println("ETHALON S:" + tt.toString());
            }
        }
        if (result_subject == null)
        {
            ExtendedIterator<Triple> it = result.find(null, Node.createURI(ns_full_rdf + "type"), Node.createURI(ns_full_msg + "Message"));

            if (it.hasNext())
            {
                Triple tt = it.next();
                result_subject = tt.getSubject();
                System.out.println(" RESULT S:" + tt.toString());
            }
        }

        if (ethalon_subject == null)
        {
            throw new Exception("in ethalon message, not found subject of type msg:Message");
        }

        if (result_subject == null)
        {
            throw new Exception("in result message, not found subject of type msg:Message");
        }
        {
            ExtendedIterator<Triple> it = ethalon.find(ethalon_subject, null, null);
            while (it.hasNext())
            {
                Triple etnalon_triple = it.next();
                ethalon_subject = etnalon_triple.getSubject();
                Node eth_value = null;
                String eth_value_str = null;
                eth_value = etnalon_triple.getObject();

                ExtendedIterator<Triple> it_res = null;

                if (etnalon_triple.getObject().isLiteral() != true)
                {
                    it_res = result.find(result_subject, etnalon_triple.getPredicate(), eth_value);
                } else
                {
                    eth_value_str = (String) eth_value.getLiteral().getValue();

                    if (eth_value_str.charAt(1) == '{' && eth_value_str.charAt(eth_value_str.length() - 2) == '}' && eth_value_str.indexOf("@") > 0 && eth_value_str.indexOf(":") > 0)
                    {
                        // найден подстановочный шаблон, извлечем данные для подстановки по правилу указанному в шаблоне
                        // src@prefix:predicate
                        // поле [src] на данный момент может иметь только значение [in], что обозначает входящее сообщение
                        eth_value_str = eth_value_str.substring(1, eth_value_str.length() - 1);

                        String[] tokens = eth_value_str.split("@");

                        if (tokens == null || tokens.length != 2 || tokens[0].equals("in") == false)
                        {
                            throw new Exception("in etnalon message, invalid template [" + eth_value + "]");
                        }

                        String[] qq = tokens[1].split(":");
                        String prefix = qq[0];

                        String template_src_value_name = prefix.replaceAll(prefix, nsShort__nsFull.get(prefix)) + qq[1];

                        ExtendedIterator<Triple> it_template_src_value = input.find(null, Node.createURI(template_src_value_name), null);

                        if (it_template_src_value.hasNext())
                        {
                            Triple tt = it_template_src_value.next();
                            // произведем подстановку
                            System.out.println("{" + eth_value + "} -> " + tt.getObject().toString());
                            eth_value_str = tt.getObject().toString();
                        }

                    }

                    it_res = result.find(result_subject, etnalon_triple.getPredicate(), Node.createLiteral(eth_value_str));
                }

                //Node.createLiteral(eth_value)
                if (it_res.hasNext())
                {
                    Triple tt = it.next();
                } else
                {
                    throw new Exception("[" + etnalon_triple + "] not found in result message");

                }


            }

        }


        return true;
    }
}
