/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package test001;

import com.hp.hpl.jena.rdf.model.Model;
import com.hp.hpl.jena.rdf.model.ModelFactory;
import com.hp.hpl.jena.rdf.model.RDFReader;
import com.hp.hpl.jena.util.FileManager;
import java.io.InputStream;

/**
 *
 * @author itiu
 */
public class Main
{

    /**
     * @param args the command line arguments
     */
    public static void main(String[] args) throws Exception
    {

        String inputFileName = "test001-in.n3";

        // create an empty model
        Model model = ModelFactory.createDefaultModel();

        // use the FileManager to find the input file
        InputStream in = FileManager.get().open(inputFileName);
        if (in == null)
        {
            throw new IllegalArgumentException(
                    "File: " + inputFileName + " not found");
        }

        RDFReader r = model.getReader("N3");

        String baseURI = "";

        r.read(model, in, baseURI) ;

        model.read(in, null);

        in.close ();

// write it to standard out
        model.write(System.out);
    }
}
