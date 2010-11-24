/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.rdf.model.ModelFactory;
import com.hp.hpl.jena.rdf.model.RDFReader;
import com.hp.hpl.jena.util.FileManager;
import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.io.StringReader;

/**
 *
 * @author itiu
 */
public class utils
{

    public static Model get_message_from_file(String inputFileName) throws Exception
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

//        message.read(in, null);

        in.close();

        return message;
    }

    public static Model get_message_from_string(String src) throws Exception
    {
        // create an empty model
        Model message = ModelFactory.createDefaultModel();

        src = Predicates.all_prefixs + src;

        StringReader sr = new StringReader (src);

//        InputStream in = new ByteArrayInputStream(src.getBytes());
//        if (in == null)
//        {
//            throw new IllegalArgumentException(
//                    "StringBufferInputStream don't created");
//        }

        RDFReader r = message.getReader("N3");

        String baseURI = "";

        r.read(message, sr, baseURI);

//        message.read(sr, null);

        sr.close();

        return message;
    }
}
