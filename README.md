```text
$ gateway -help
gateway is a basic dynamic webserver that delegates handling of HTTP requests 
to executables on disk using the Common Gateway Interface (CGI).

For example, an executable with the basename qux is responsible for handling a 
request for the /qux HTTP resource. An executable with the special basename _ 
is responsible for handling a request for the / HTTP resource (i.e. the 
homepage).

An executable handles a request by writing HTTP headers (Status, Content-Type, 
...) followed by some content (likely HTML) to standard output, and then 
exiting. CGI information is conveyed to executables using environment variables 
with standard names - see http://www.cgi101.com/book/ch3/text.html

usage:
  gateway [flags] /path/to/executables/dir

flags:
  -copyenv
        include a copy of the server's own environment variables
  -port int
        tcp port number on which to listen for connections (default 80)
  -wildcard
        have _ perform double-duty and also handle any HTTP resource that isn't 
otherwise handled
```
