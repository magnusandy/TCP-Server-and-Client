package main

import "net"
import "fmt"
import "bufio"
import "os"

//Handles the input sent back to the client from the server, simply writes it to the console
//
func getFromServer(conn net.Conn){
  reader := bufio.NewReader(conn)
  for{
    message, _ := reader.ReadString('\n')
    fmt.Print(message)
  }
}

//Handles user input, reads from stdin and then posts that line to the server
func getfromUser(conn net.Conn){
    for{
      reader := bufio.NewReader(os.Stdin)
      text, _ := reader.ReadString('\n')
      fmt.Fprintf(conn, text)
    }
  }

//starts up the client, starts the recieving thread and the input threads and then loops forever
func main() {
  // connect to this socket
  conn, _ := net.Dial("tcp", "45.55.235.148:25563")
  go getFromServer(conn);
  go getfromUser(conn);
  for{}
}
