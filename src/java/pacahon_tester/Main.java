/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.graph.Graph;
import com.hp.hpl.jena.rdf.model.Model;
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

    //  private static Predicates predicates;
    public static void main(String[] args) throws Exception
    {
        //      predicates = new Predicates();
        /////////////////////////////////////////////////////////////////////////////

        String defaultConnectTo="tcp://127.0.0.1:5555";
        String connectTo = args[0];
//        String connectTo = "ipc://worker";
        ZMQ.Context ctx = ZMQ.context(1);
        ZMQ.Socket socket = ctx.socket(ZMQ.REQ);

        socket.connect(connectTo);

        /////////////////////////////////////////////////////////////////////////////

        for (int ii = 1; ii < args.length; ii++)
        {
            String test_name = args[ii];
            System.out.println("\nTEST:" + test_name);

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
                    out.write(Predicates.all_prefixs + "\n" + result);
                    out.close();

                    Tester tt = new Tester();

                    if (tt.cmp_results(input_message, result_message, output_message, null, null) == false)
                    {
                        throw new Exception("result != ethalon");
                    } else
                    {
                        System.out.println("test [" + test_name + "] is passed");
                    }
                } catch (Exception ex)
                {
                    System.out.println("RES:\n" + result_message);
                    System.out.println("ETHALON:\n" + output_message);
                    throw ex;
                }

            }

            long end = System.currentTimeMillis();
            System.out.println("total time: (" + (end - start) + "[ms])");
        }
    }
    //
}
