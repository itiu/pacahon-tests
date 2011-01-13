/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package pacahon_tester;

import java.util.HashMap;

/**
 *
 * @author itiu
 */
public class Predicates
{

    public final static String ns_full_rdf = "http://www.w3.org/1999/02/22-rdf-syntax-ns#";
    public final static String ns_full_msg = "http://gost19.org/message#";
    public final static String ns_full_auth = "http://gost19.org/auth#";
    public final static String ns_full_rdfs = "http://www.w3.org/2000/01/rdf-schema#";
    public final static String ns_full_xsd = "http://www.w3.org/2001/XMLSchema#";
    //
    public final static String prefix_rdf = "@prefix rdf:     <" + ns_full_rdf + "> .";
    public final static String prefix_rdfs = "@prefix rdfs:    <" + ns_full_rdfs + "> .";
    public final static String prefix_xsd = "@prefix xsd:     <" + ns_full_xsd + "> .";
    public final static String prefix_msg = "@prefix msg:     <" + ns_full_msg + "> .";
    public final static String prefix_auth = "@prefix auth:     <" + ns_full_auth + "> .";
    //
    public final static String all_prefixs = prefix_rdf + "\n" + prefix_rdfs + "\n" + prefix_xsd + "\n" + prefix_msg + "\n" + prefix_auth + "\n";
    //
    public HashMap<String, String> nsShort__nsFull;

    Predicates()
    {
        nsShort__nsFull = new HashMap<String, String>();

        nsShort__nsFull.put("rdf", Predicates.ns_full_rdf);
        nsShort__nsFull.put("msg", Predicates.ns_full_msg);
        nsShort__nsFull.put("auth", Predicates.ns_full_auth);
    }
}
