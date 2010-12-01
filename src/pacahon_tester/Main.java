/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.graph.Graph;
import com.hp.hpl.jena.graph.Node;
import com.hp.hpl.jena.graph.Triple;
import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.util.iterator.ExtendedIterator;
import java.io.BufferedWriter;
import java.io.ByteArrayOutputStream;
import java.io.FileWriter;
import org.zeromq.ZMQ;

/**
 *
 * @author itiu
 */
public class Main
{

    public static void main(String[] args) throws Exception
    {

        String test_name = "test001";

        if (args.length > 0)
        {
            test_name = args[0];
        }

        System.out.println("TEST:" + test_name);
        /////////////////////////////////////////////////////////////////////////////

        String connectTo = "tcp://127.0.0.1:5555";
//        String connectTo = "ipc://worker";
        ZMQ.Context ctx = ZMQ.context(1);
        ZMQ.Socket socket = ctx.socket(ZMQ.REQ);

        socket.connect(connectTo);

        /////////////////////////////////////////////////////////////////////////////

        Model m_input_message = utils.get_message_from_file(test_name + "-in.n3");
        Graph input_message = m_input_message.getGraph();
        Graph output_message = utils.get_message_from_file(test_name + "-out.n3").getGraph();

        Graph result_message = null;

        ByteArrayOutputStream baos = new ByteArrayOutputStream();
        m_input_message.write(baos, "N3");
        byte data[] = baos.toByteArray();

        long start = System.currentTimeMillis();


        String result = null;

        for (int i = 0; i < 1; i++)
        {
            long start_server_io = System.nanoTime();

            System.out.println("SEND COMMAND TO SERVER");

            socket.send(data, 0);

            long start_recieve_time = System.nanoTime();

            byte[] rr = socket.recv(0);
            long end_server_io = System.nanoTime();

            System.out.println("server i/o time : (" + (end_server_io - start_server_io) / 1000 + "[µs], recieve time = " + (end_server_io - start_recieve_time) / 1000 + "[µs])");

            result = new String(rr);

            System.out.println("OUT: \n" + result);

            // сравниваем результаты полученные и результаты из шаблона xxxx-out.n3

            try
            {
                result_message = utils.get_message_from_string(result).getGraph();

                BufferedWriter out = new BufferedWriter(new FileWriter(test_name + "-recieve.n3"));
                out.write(Predicates.all_prefixs + result);
                out.close();

                if (cmp_results(input_message, result_message, output_message, null, null) == false)
                {
                    throw new Exception("result != ethalon");
                } else
                {
                    System.out.println("test [" + test_name + "] is passed");
                }
            } catch (Exception ex)
            {
                System.out.println("RES:\n" + result_message);
                throw ex;
            }

        }

        long end = System.currentTimeMillis();
        System.out.println("total time: (" + (end - start) + "[ms])");

    }
    private static Node input_subject = null;


    /*
     * сравнение графа шаблона и графа результата
     */
    public static boolean cmp_results(Graph input, Graph result, Graph ethalon, Node cur_result_subject, Node cur_ethalon_subject) throws Exception
    {
        if (cur_ethalon_subject == null)
        {
            ExtendedIterator<Triple> it = ethalon.find(null, Node.createURI(Predicates.ns_full_rdf + "type"), Node.createURI(Predicates.ns_full_msg + "Message"));

            if (it.hasNext())
            {
                Triple tt = it.next();
                cur_ethalon_subject = tt.getSubject();
                System.out.println("ETHALON S:" + tt.toString());
            }
        }
        if (cur_result_subject == null)
        {
            ExtendedIterator<Triple> it = result.find(null, Node.createURI(Predicates.ns_full_rdf + "type"), Node.createURI(Predicates.ns_full_msg + "Message"));

            if (it.hasNext())
            {
                Triple tt = it.next();
                cur_result_subject = tt.getSubject();
                System.out.println(" RESULT S:" + tt.toString());
            }
        }
        if (input_subject == null)
        {
            ExtendedIterator<Triple> it = input.find(null, Node.createURI(Predicates.ns_full_rdf + "type"), Node.createURI(Predicates.ns_full_msg + "Message"));

            if (it.hasNext())
            {
                Triple tt = it.next();
                input_subject = tt.getSubject();
                System.out.println(" INPUT S:" + tt.toString());
            }
        }

        if (input_subject == null)
        {
            throw new Exception("in input message, not found subject of type msg:Message");
        }

        if (cur_ethalon_subject == null)
        {
            throw new Exception("in ethalon message, not found subject of type msg:Message");
        }

        if (cur_result_subject == null)
        {
            throw new Exception("in result message, not found subject of type msg:Message");
        }


        ExtendedIterator<Triple> it = ethalon.find(cur_ethalon_subject, null, null);
        while (it.hasNext())
        {
            Triple etnalon_triple = it.next();
            cur_ethalon_subject = etnalon_triple.getSubject();
            Node eth_value = null;
            String eth_value_str = null;
            eth_value = etnalon_triple.getObject();

            if (etnalon_triple.getObject().isLiteral() == false)
            {
                ExtendedIterator<Triple> it_res = null;

                it_res = result.find(cur_result_subject, etnalon_triple.getPredicate(), null);
                if (it_res != null && it_res.hasNext())
                {
                    Triple tt = it_res.next();

                    cmp_results(input, result, ethalon, tt.getObject(), etnalon_triple.getObject());

                } else
                {
                    throw new Exception("[" + etnalon_triple.getPredicate() + "] not found in result message");

                }

            }

            if (etnalon_triple.getObject().isLiteral() == true)
            {
                eth_value_str = (String) eth_value.getLiteral().getValue();

                if (eth_value_str.charAt(0) == '{' && eth_value_str.charAt(eth_value_str.length() - 1) == '}' && eth_value_str.indexOf("@") > 0 && eth_value_str.indexOf(":") > 0)
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

                    if (qq[1].equals("subject"))
                    {
                        eth_value = input_subject;
                    } else
                    {
                        String template_src_value_name = prefix.replaceAll(prefix, Predicates.nsShort__nsFull.get(prefix)) + qq[1];

                        ExtendedIterator<Triple> it_template_src_value = input.find(null, Node.createURI(template_src_value_name), null);

                        if (it_template_src_value.hasNext())
                        {
                            Triple tt = it_template_src_value.next();
                            // произведем подстановку
                            System.out.println("{" + eth_value + "} -> " + tt.getObject().toString());
                            eth_value_str = (String) tt.getObject().getLiteralValue();
                        } else
                        {
                            throw new Exception("[" + etnalon_triple.getPredicate() + "][" + eth_value + "] value for replace not found");
                        }
                        eth_value = Node.createLiteral(eth_value_str);
                    }

                }

                ExtendedIterator<Triple> it_res = null;

                it_res = result.find(cur_result_subject, etnalon_triple.getPredicate(), null);
                boolean isFound = false;
                while (it_res.hasNext())
                {
                    Triple tt = it_res.next();

                    Node res_value = tt.getObject();

                    if (eth_value.isLiteral())
                    {
                        String str_eth_value = (String) eth_value.getLiteralValue();
                        String str_res_value = (String) res_value.getLiteralValue();

                        if (str_eth_value.equals(str_res_value) == true)
                        {
                            isFound = true;
                        }

                        if (isFound == false)
                        {
                            isFound = true;
                            for (int i = 0; i < str_eth_value.length(); i++)
                            {
                                if (str_eth_value.charAt(i) != str_res_value.charAt(i) && str_eth_value.charAt(i) != '$')
                                {
                                    isFound = false;
                                    break;
                                }
                            }
                        }

                    } else
                    {
                        if (eth_value.equals(res_value))
                        {
                            isFound = true;
                        }
                    }
                }

                if (isFound == false)
                {
                    throw new Exception("[" + etnalon_triple.getPredicate() + "][" + eth_value + "] not found in result message");

                }
            }

        }

        return true;
    }
}
