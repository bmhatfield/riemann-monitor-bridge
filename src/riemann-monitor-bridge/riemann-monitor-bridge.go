package main

// Stdlib
import "os"
import "log"
import "time"
import "flag"
import "runtime"
import "net/http"

// 3rd Party
import "github.com/amir/raidman"

const VERSION = "0.0.1"

var showVersion *bool = flag.Bool("version", false, "Print version to screen and exit")

var riemannHost *string = flag.String("host", "localhost:5555", "Host:Port combination")

func pingHandler(resp http.ResponseWriter, req *http.Request) {
    c, err := raidman.Dial("tcp", *riemannHost)
    if err != nil {
        log.Println(err)
        resp.WriteHeader(500)
        resp.Write([]byte("Unable to connect to Riemann"))
        return
    } else {
        defer c.Close()
    }

    var event = &raidman.Event{
        State:   "ok",
        Host:    "http_pinger",
        Service: "monitoring-bridge",
        Ttl:     600,
        Tags:    []string{"nonotify"},
    }

    err = c.Send(event)
    if err != nil {
        log.Println(err)
        resp.WriteHeader(500)
        resp.Write([]byte("Unable to send event to Riemann"))
        return
    }

    events, err := c.Query("(service = \"monitoring-bridge\")")
    if err != nil {
        log.Println(err)
        resp.WriteHeader(500)
        resp.Write([]byte("Unable to query event from Riemann"))
        return
    }

    if len(events) < 1 {
        resp.WriteHeader(500)
        resp.Write([]byte("Event written but not found"))
        return
    } else {
        log.Printf("Event Found: %+v", events[0])
        resp.Write([]byte("Event Found!"))
    }
}

func main() {
    // We want to use every ounce of computing power, so let's switch on
    // every one of our cores! The docs say this call will go away in the
    // future, so future compilers may have to remove this line.
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Parse our flags! Note, flags may be defined *anywhere* in a Go program,
    // so even if there isn't one in our main program, that doesn't mean other
    // libraries aren't hoping to make use of some optional behavior.
    flag.Parse()

    // If it's just been requested that we print a version string, print that
    // and exit. This depends on flag.Parse() having been run already.
    if *showVersion {
        log.Printf("Riemann-Monitor-Bridge %s\n", VERSION)
        os.Exit(0)
    }

    http.HandleFunc("/ping", pingHandler)

    // Log that we've started
    log.Printf("Riemann-Monitor-Bridge %s configured and preparing to serve...", VERSION)

    // And start the HTTP server!
    server := &http.Server{
        Addr:           ":20080",
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 21,
    }

    err := server.ListenAndServe()

    if err != nil {
        log.Printf("Unable to start HTTP server: %s", err)
    }

    // This will not usually get called, as ListenAndServe will block forever.
    // However, there are some circumstances in which ListenAndServe will stop,
    // in which case Main() will complete after printing this log message.
    log.Printf("Riemann-Monitor-Bridge %s exiting!", VERSION)
}
