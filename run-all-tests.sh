java -Djava.library.path=/usr/local/lib -jar "dist/tester.jar" tests/test001
java -Djava.library.path=/usr/local/lib -jar "dist/tester.jar" tests/test002
java -Djava.library.path=/usr/local/lib -jar "dist/tester.jar" tests/test003
ticket=`grep ticket tests/test003-recieve.n3 | cut -c 20-`
echo $ticket
java -Djava.library.path=/usr/local/lib -jar "dist/tester.jar" tests/test004
sed -e "12 c msg:ticket $ticket ;" tests/test005-in.n3-src > tests/test005-in.n3
java -Djava.library.path=/usr/local/lib -jar "dist/tester.jar" tests/test005
