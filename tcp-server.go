package main

import "net"
import "fmt"
import "bufio"
import "./myUtils"
import "time"
import "strings"

//CONSTANTS
const SERVER_IP string = "localhost";
const SERVER_PORT string = "8080";
const COMMAND_PREFIX string = "/";

//COMMANDS
const HELP_COMMAND string = COMMAND_PREFIX+"help";
const QUIT_COMMAND string = COMMAND_PREFIX+"quit";
const CREATE_ROOM_COMMAND string = COMMAND_PREFIX+"createRoom"; //creates a room with the name of the first argument given
const LIST_ROOMS_COMMAND string = COMMAND_PREFIX+"listRooms"
const JOIN_ROOM_COMMAND string = COMMAND_PREFIX+"join";//   /join roomname will add a user to a rooms list of clients and switch the user to that room
const CURR_ROOM_COMMAND string = COMMAND_PREFIX+"currentRoom";
var HELP_INFO = [...]string {"help and command info: ",
 HELP_COMMAND+": use this command to get some help",
 QUIT_COMMAND+": Safely exit the system",
 CREATE_ROOM_COMMAND+" roomName : creates a room with the name roomName",
 LIST_ROOMS_COMMAND+": lists all rooms available for joining",
}
var MessageStorageArray []string;
var connectionArray []net.Conn;
var ClientArray []Client;
var RoomArray []Room;
//STRUCTURES
/*****************Rooms*****************/
type Room struct{
  name string;
  clientList []*Client;
  createdDate time.Time;
  lastUsedDate time.Time;
  chatLog []ChatMessage;
  creator *Client;
}

func createRoom(roomName string, roomCreator *Client) Room {
  var newRoom = Room{
    name: roomName,
    clientList: make([]*Client, 0),//room will start empty, we wont add the creator in
    createdDate: time.Now(),
    lastUsedDate: time.Now(),
    chatLog: make([]ChatMessage, 0),
    creator: roomCreator,
  }
  RoomArray = append(RoomArray, newRoom);
  return newRoom;
}

//returns true if a user is already in the room, false otherwise
func (room Room) isClientInRoom(client *Client) bool {
  for _, roomClient := range room.clientList {
    if client.name == roomClient.name {
      return true;
    }
  }
  return false;
}
/***************************************/

/*****************MESSAGES*****************/

//Structure holding messages sent to a chat, stores meta information on the client who sent it
type ChatMessage struct {
  client Client;
  message string;
  createdDate time.Time;
}

//creates a new instance of a ChatMessage and returns it
func createChatMessage(cli Client, mess string) ChatMessage {
 var chatMessage = ChatMessage{
   client: cli,
   message: mess,
   createdDate: time.Now(),
 }
 return chatMessage;
}
/******************************************/

/*****************CLIENTS*****************/
//Clients have names, and a reader and writer as well as a link to their connection
//Client names are garenteed by the generateName fucntion to be unique for the duratoin of program execution (NOT persisted)
type Client struct
{
  connection net.Conn;
  readListener bufio.Reader;
  writeListener bufio.Writer;
  currentRoom *Room;
  outputChannel chan string;
  name string;
}

//this funciton watches the clients output channel, when something is added to the channel,
func (cli *Client) WaitForAWrite(){
  //looping forever
    //loop watching the clients output channel
    for output := range cli.outputChannel {
      if cli.connection == nil {
        return;
      }
      _, error := cli.writeListener.WriteString(output)
      if error != nil{
        fmt.Println(error)
        break
      }
      fmt.Println(output)
      //flushing is necessary, the writeString only takes in the string, the flush function pushes it out to the user
      flushError := cli.writeListener.Flush()
      if flushError != nil {
        fmt.Println(flushError)
        break
      }
    }
}

//adds message to the clients output channel, messages should be single line, NON delimited strings, that is the message should not include a new line
func (cli Client) messageClient(message string){
  message = message+"\n";
  cli.outputChannel <- message;
}


//Intended to be run on a thread, this function will wait and lisen for messages from the client
func (cli *Client)WaitForARead(){
  for{
    message, _ := cli.readListener.ReadString('\n')
    fmt.Print("Message Received:", string(message))
    checkForCommand(message, cli);
    //WriteToAllChans(message, cli);
  }
}
/**********************************/

/*
Takes in a connection and creates a Client for it,
 adds a read and write listener and starts them on seperate GO threads
 as well as opens the output chan
*/
func addClient(conn net.Conn){
   createReader := bufio.NewReader(conn);
   createWriter := bufio.NewWriter(conn);
   createOutputChannel := make(chan string);
   createName := myUtils.GenerateName();

    var cli  = Client{
    connection: conn,
    readListener: *createReader,
    writeListener: *createWriter,
    currentRoom: nil, //starts as nil because the user is not initally in a room
    outputChannel: createOutputChannel,
    name: createName,
  }

  ClientArray = append(ClientArray, cli);
  defer cli.messageClient("Welcome to the Server, Your username for this session is: "+cli.name);
  go cli.WaitForARead();
  go cli.WaitForAWrite();
}



//writes to all the channels of all the users but the one that posts it, to avoid double posting
func WriteToAllChans(message string, senderClient *Client){
  for i := range ClientArray {
    if senderClient.connection != ClientArray[i].connection{
      ClientArray[i].messageClient(message);
    }
  }
}

/*
Checks if the line sent from the user includes a command
Commands will be in the form of /Command arg
this function will first check if the FIRST character of the clients string is a /,
if it is then it will attempt to parse and execute the command.
*/
func checkForCommand(message string, client *Client) {
  message = strings.TrimSpace(message);//strips the newlines from the string
  isCommand := strings.HasPrefix(message, COMMAND_PREFIX);//checks to see if the line starts with /
  if(isCommand){
    //parse command line, commands should be in the exact form of "/command arg arg arg" where args are not required
    parsedCommand := strings.Split(message, " ")
    if parsedCommand[0] == HELP_COMMAND {
       processHelpCommand(client);
    } else if parsedCommand[0] == QUIT_COMMAND {
      processQuitCommand(client);
    } else if parsedCommand[0] == CREATE_ROOM_COMMAND {
      processCreateRoomCommand(client, parsedCommand[1]);//TODO handle the possibility that the user doesnt add an arg just types "/createRoom"
    } else if parsedCommand[0] == LIST_ROOMS_COMMAND {
      processListRoomsCommand(client);
    } else if parsedCommand[0] == JOIN_ROOM_COMMAND {
      processJoinRoomCommand(client, parsedCommand[1]);
    } else if parsedCommand[0] == CURR_ROOM_COMMAND {
      processCurrRoomCommand(client);
    }

  } else { // message is not a command
    WriteToAllChans(message, client);
  }
}
 func processCurrRoomCommand (client *Client){
   client.messageClient("current room: "+client.currentRoom.name);
 }

//Loops through the HELP_INFO array and sends all the lines of help info to the user
func processHelpCommand(client *Client){
       for _, helpLine := range HELP_INFO{
         client.messageClient(helpLine);
       }
}

//quits the client from the server
func processQuitCommand(client *Client){
  client.connection.Close();
  client.connection = nil;
}

//creates a room and logs to the console
func processCreateRoomCommand(client *Client, roomName string){
  room := createRoom(roomName, client);
  fmt.Println(room.creator.name+" created a room called: "+room.name)
}

//sends the list of rooms to the client
func processListRoomsCommand(client *Client){
  client.messageClient("List of rooms:")
  for _, roomName := range RoomArray{
    client.messageClient(roomName.name);
  }
  client.messageClient("");
}

//returns true of the room was joined successfully, returns false if there was a problem like the room does not exist
func processJoinRoomCommand(client *Client, roomName string) bool{
  //start by checking if the room exists
  roomToJoin := getRoomByName(roomName);
  if roomToJoin == nil{ //the room doesnt exist
    fmt.Println(client.name+" tried to enter room: "+roomName+" which does not exist");
    return false;
  }
  //Room exists so now we can join it.
  //check if user is already in the room
  //add user to room if not in it already
  if roomToJoin.isClientInRoom(client) {
    //already in there so no worries
  } else {
    roomToJoin.clientList = append(roomToJoin.clientList, client);// add client to the rooms list
  }
  //switch users current room to room
  client.currentRoom = roomToJoin;
  fmt.Println(client.currentRoom.name)
  //display all messages in the room
  //TODO
  return true
}

//checks to see if a room with the given name exists in the RoomArray, if it does return it, if not return nil
func getRoomByName(roomName string) *Room{
  for _, room := range RoomArray{
    if room.name == roomName{
      return &room;
    }
  }
  return nil;
}



//Main function for starting the server, will open the server on the SERVER_IP and the SERVER_PORT
func main() {
  fmt.Println("Launching server...")
  //Start the server on the constant IP and port
  ln, connectError := net.Listen("tcp", SERVER_IP+":"+SERVER_PORT)
  //check for errors in the server starup
  if connectError != nil {
    fmt.Println("Error Launching server "+ connectError.Error())
  }else{
    fmt.Println("Server Started")
  }
  //Initialize the connectionArray, this will hold all the incoming connections
  connectionArray = make([]net.Conn, 0);
  // run loop forever, accept connections when they come and add them to the connection array and then call the addClient function one
  for {
    conn, _ := ln.Accept()
    connectionArray = append(connectionArray, conn)
    addClient(conn);
  }
}
