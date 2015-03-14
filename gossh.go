package main

import "os"
import "fmt"
import "log"
import "flag"
import "bytes"
import "strings"
import "os/user"
import "io/ioutil"
import "golang.org/x/crypto/ssh"

var flag_workers = flag.Int("workers", 3, "number of workers")
var flag_identity = flag.String("identity", "", "path to sshkey")
var flag_username = flag.String("username", "", "ssh username")

func init() {
    // setup flags / usage
    usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }
    flag.Usage = func() {
        fmt.Println("\nUsage:")
        fmt.Println("cat hostlist | gossh [-w|--workers] [-u|--username] [-i|--identity] 'cmd to run'")
        fmt.Println("  --workers -w -- Number of workers to spawn (default: 3)")
        fmt.Println("  --username -u -- Username to use for ssh connections (default: " + usr.Username + ")")
        fmt.Println("  --identity -i -- ssh private key to use (default: " + usr.HomeDir + "/.ssh/id_rsa)")
        fmt.Println("\nhostlist must be \\n delimited list\n")
    }
    flag.IntVar(flag_workers, "w", 3, "number of workers")
    flag.StringVar(flag_identity, "i", "", "path to sshkey")
    flag.StringVar(flag_username, "u", "", "ssh username")
    flag.Parse()
    if *flag_identity == "" {
        *flag_identity = usr.HomeDir + "/.ssh/id_rsa"
    }
    if *flag_username == "" {
        *flag_username = usr.Username
    }
    if len(flag.Args()) < 1 {
        flag.Usage()
        os.Exit(1)
    }
}

func main() {
    // check for stdin over pipe
    fi, err := os.Stdin.Stat()
    if err != nil {
        panic(err)
    }
    if fi.Mode() & os.ModeNamedPipe == 0 {
        flag.Usage()
        os.Exit(1)
    }

    // read from stdin
    bytes, err  := ioutil.ReadAll(os.Stdin)
    if err != nil {
        panic(err)
    }

    // split stdin on \n
    var hosts []string
    hosts = strings.Split(string(bytes), "\n")

    // setup channels
    jobs := make(chan string, 100)
    results := make(chan string, 100)

    // setup workers
    for w:= 1; w <= *flag_workers; w++ {
        go ssh_worker(w, jobs, results)
    }

    // send jobs
    var job_count int
    job_count = 0
    for j :=0; j<=len(hosts)-1; j++ {
        if hosts[j] != "" {
            jobs <- hosts[j]
            job_count += 1
        }
    }
    close(jobs)

    // print results
    for a := 0; a <= job_count-1; a++ {
       <-results
    }    
}

func ssh_worker(id int, jobs <- chan string, results chan<- string) {
    for j := range jobs {
        fmt.Println("ssh_worker", id, "processing host", j)
        results <- ssh_connect(j, "22", *flag_username, strings.Join(flag.Args(), " "))
    }
}

func ssh_connect(ip string, port string, username string, cmd_line string) (string) {
    pkey := parsekey(*flag_identity)

    config := &ssh.ClientConfig{
        User: username,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(pkey),
        },
    }
    client, err := ssh.Dial("tcp", ip+":"+port, config)
    if err != nil {
        panic("Failed to dial: " + err.Error())
    }

    // Each ClientConn can support multiple interactive sessions,
    // represented by a Session.
    session, err := client.NewSession()
    if err != nil {
        panic("Failed to create session: " + err.Error())
    }
    defer session.Close()

    // Once a Session is created, you can execute a single command on
    // the remote side using the Run method.
    var b bytes.Buffer
    session.Stdout = &b
    if err := session.Run(cmd_line); err != nil {
        //panic("Failed to run: " + err.Error())
        fmt.Println("failed to session.run:" + err.Error())
        fmt.Println("returned:" + b.String())
        return ""
    }
    fmt.Println(b.String())
    return b.String()
}

func parsekey(file string) ssh.Signer {
    privateBytes, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Println(err)
        panic("Failed to load private key")
    }

    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        panic("Failed to parse private key")
    }
    return private
}
