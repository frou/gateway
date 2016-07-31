package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/frou/stdext"
)

const (
	rootResource         = "/"
	rootResourceExecName = "_"

	wildcardResourceRepr = "*"
)

var (
	port = flag.Int("port", 80,
		"tcp port number on which to listen for connections")

	copyEnv = flag.Bool("copyenv", false,
		"child processes get a copy of the server's environment variables")

	withEnv = flag.String("withenv", "",
		"child processes get exactly these environment variables, "+
			"specified in the form k0=v0,k1=v1,...")

	wildcard = flag.Bool("wildcard", false,
		"have "+rootResourceExecName+" perform double-duty and also handle "+
			"any HTTP resource that isn't otherwise handled")

	mappingWriter = tabwriter.NewWriter(os.Stderr, 0, 8, 1, ' ', 0)
)

func main() {
	stdext.SetPreFlagsUsageMessage(desc, "/path/to/executables/dir")
	if err := stdext.ParseFlagsExpectingNArgs(1); err != nil {
		stdext.Exit(err)
	}

	execPaths, err := findExecPaths(flag.Arg(0))
	if err != nil {
		stdext.Exit(err)
	}

	setupHandlers(execPaths)
	stdext.Exit(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func findExecPaths(dir string) ([]string, error) {
	execDir, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer execDir.Close()
	execInfos, err := execDir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	var execPaths []string
	for _, execInfo := range execInfos {
		if !execInfo.Mode().IsRegular() {
			continue
		}
		execPaths = append(
			execPaths,
			filepath.Join(execDir.Name(), execInfo.Name()))
	}
	return execPaths, nil
}

func setupHandlers(execPaths []string) {
	var childEnv []string
	if *withEnv != "" {
		for _, kvp := range strings.Split(*withEnv, ",") {
			childEnv = append(childEnv, kvp)
		}
	} else if *copyEnv {
		childEnv = os.Environ()
	}
	defer mappingWriter.Flush()
	for _, execPath := range execPaths {
		cgiHandler := &cgi.Handler{
			Path: execPath,
			Dir:  ".",
			Env:  childEnv,
		}
		resource := rootResource
		execName := filepath.Base(execPath)
		if execName == rootResourceExecName {
			if *wildcard {
				http.Handle(resource, cgiHandler)
				defer printMapping(wildcardResourceRepr, execPath)
			} else {
				http.HandleFunc(resource,
					func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path != resource {
							http.NotFound(w, r)
							return
						}
						cgiHandler.ServeHTTP(w, r)
					})
			}
		} else {
			resource += execName
			http.Handle(resource, cgiHandler)
		}
		printMapping(resource, execPath)
	}
}

func printMapping(resource, execPath string) {
	fmt.Fprintf(mappingWriter, "%s\t->\t%s\n", resource, execPath)
}

// TODO: Don't have this weird embedded readme in the code or a copy-pasted
// invocation into the readme file.
var desc = fmt.Sprintf(`
%s is a basic dynamic webserver that delegates handling of HTTP requests to
executables on disk using the Common Gateway Interface (CGI).

For example, an executable with the basename qux is responsible for handling
a request for the /qux HTTP resource. An executable with the special basename
%s is responsible for handling a request for the / HTTP resource (i.e. the
homepage).

An executable handles a request by writing HTTP headers (Status, Content-Type,
...) followed by some content (likely HTML) to standard output, and then
exiting. CGI information is conveyed to executables using environment
variables with standard names - see http://www.cgi101.com/book/ch3/text.html

Go's standard library does the heavy-lifting.
`, os.Args[0], rootResourceExecName)
