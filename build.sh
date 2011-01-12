mkdir dist
javac -cp libs/icu4j-3.4.4.jar:libs/iri-0.8.jar:libs/jena-2.6.3.jar:libs/log4j-1.2.13.jar:libs/slf4j-api-1.5.8.jar:libs/slf4j-log4j12-1.5.8.jar:libs/xercesImpl-2.7.1.jar:libs/zmq.jar \
src/java/pacahon_tester/*.java -d dist
cp src/java/myManifest dist/myManifest
cd dist
jar cfm tester.jar myManifest pacahon_tester/*.class
cd ..
cp dist/tester.jar tester.jar
