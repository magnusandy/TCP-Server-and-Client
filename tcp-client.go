package main

import "net"
import "fmt"
import "bufio"
import "os"
import "strings"

var stayAlive bool = true;

//Handles the input sent back to the client from the server, simply writes it to the console
//
func getFromServer(conn net.Conn){
  reader := bufio.NewReader(conn)
  for{
    message, _ := reader.ReadString('\n')
    if message == "SERVER FULL"{
      fmt.Println("Server is full, please try again later.")
      stayAlive = false;
      return;
    } else if message == "Server says: TIMEOUT\n" {
      fmt.Println("You timed out, please reconnect")
      stayAlive = false;
      return;
    }
    fmt.Print(message)
  }
}

//Handles user input, reads from stdin and then posts that line to the server
func getfromUser(conn net.Conn){
    for{
      reader := bufio.NewReader(os.Stdin)
      text, _ := reader.ReadString('\n')

      fmt.Fprintf(conn, text)
      if strings.TrimSpace(text) == "/quit"{
        stayAlive = false;
      }
    }
  }

//starts up the client, starts the recieving thread and the input threads and then loops forever
func main() {

arguments := os.Args[1:];
IP := "localhost";
PORT:= "8080";
if len(arguments) == 0 {
  //no arguments start on localhost 8080
} else if len(arguments) != 2 {
  fmt.Println("I cannot understand your arguments, you must specify no arguments or exactly 2, first the IP and the second as the port")
  return
} else if len(arguments) == 2 {
//correct ammount of args
IP = arguments[0]
PORT = arguments[1]
}
//fmt.Println(arg)
  // connect to this socket
  fmt.Println("Attempting to connect to "+IP+":"+PORT)
  conn, err := net.Dial("tcp", IP+":"+PORT)
  if err != nil{
    fmt.Println("Something went wrong with the connection, check that the server exists and that your IP/Port are correct:\nError Message: ")
    fmt.Println(err)
    return
  }

  go getFromServer(conn);
  go getfromUser(conn);
  for stayAlive {
    //loops  forever until stayAlive is set to false and then it shuts down
  }
}
