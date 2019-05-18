package twitchbot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
	"twitchbot/db"
	"twitchbot/models"
	"twitchbot/repo"

	rgb "github.com/foresthoffman/rgblog"
)

const PSTFormat = "Jan 2 15:04:05 PST"

// Regex for parsing PRIVMSG strings.
var msgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

var cmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

type OAuthCred struct {
	Password string `json:"password,omitempty"`
}

type ITwitchBot interface {
	Connect()
	Disconnect()
	HandleChat() error
	JoinChannel()
	ReadCredentials() error
	Say(msg string) error
	Start()
}

type TwitchBot struct {
	Channel     string
	conn        net.Conn
	Credentials *OAuthCred
	MsgRate     time.Duration
	Name        string
	Port        string
	PrivatePath string
	Server      string
	startTime   time.Time
}

// Connects the bot to the Twitch IRC server.
func (bb *TwitchBot) Connect() {
	var err error
	rgb.YPrintf("[%s] Connecting to %s...\n", timeStamp(), bb.Server)

	// makes connection to Twitch IRC server
	bb.conn, err = net.Dial("tcp", bb.Server+":"+bb.Port)
	if nil != err {
		rgb.YPrintf("[%s] Cannot connect to %s, retrying.\n", timeStamp(), bb.Server)
		bb.Connect()
		return
	}
	rgb.YPrintf("[%s] Connected to %s!\n", timeStamp(), bb.Server)
	bb.startTime = time.Now()
}

//Disconnects
func (bb *TwitchBot) Disconnect() {
	bb.conn.Close()
	upTime := time.Now().Sub(bb.startTime).Seconds()
	rgb.YPrintf("[%s] Closed connection from %s! | Live for: %fs\n", timeStamp(), bb.Server, upTime)
}

// Listens for and logs messages from chat
func (bb *TwitchBot) HandleChat() error {
	rgb.YPrintf("[%s] Watching #%s...\n", timeStamp(), bb.Channel)

	// reads from connection
	tp := textproto.NewReader(bufio.NewReader(bb.conn))
	for {
		line, err := tp.ReadLine()
		if nil != err {
			bb.Disconnect()
			return errors.New("bb.Bot.HandleChat: Failed to read line from channel. Disconnected.")
		}
		rgb.YPrintf("[%s] %s\n", timeStamp(), line)

		if "PING :tmi.twitch.tv" == line {
			bb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
			continue
		} else {

			// handle a PRIVMSG message
			matches := msgRegex.FindStringSubmatch(line)
			if nil != matches {
				userName := matches[1]
				msgType := matches[2]

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					rgb.GPrintf("[%s] %s: %s\n", timeStamp(), userName, msg)
					var user = &models.User{Name: userName}
					result, err := repo.GetByName(user.Name)

					if err == nil {
						repo.Create(user)
						rgb.GPrintf("[%s] user %s created in DB\n", timeStamp(), userName)
					}
					if result != nil {
						rgb.GPrintf("[%s] %s: already in DB\n", timeStamp(), userName)
					}
					// parse commands from user message
					cmdMatches := cmdRegex.FindStringSubmatch(msg)
					if nil != cmdMatches {
						cmd := cmdMatches[1]
						// channel-owner specific commands
						if userName == bb.Channel {
							switch cmd {
							case "tbdown":
								rgb.CPrintf(
									"[%s] Shutdown command received. Shutting down now...\n",
									timeStamp(),
								)
								bb.Disconnect()
								return nil
							default:
								// do nothing
							}
						}
					}
				default:
					// do nothing
				}
			}
		}
		time.Sleep(bb.MsgRate)
	}
}

// Makes the bot join its pre-specified channel.
func (bb *TwitchBot) JoinChannel() {
	rgb.YPrintf("[%s] Joining #%s...\n", timeStamp(), bb.Channel)
	bb.conn.Write([]byte("PASS " + bb.Credentials.Password + "\r\n"))
	bb.conn.Write([]byte("NICK " + bb.Name + "\r\n"))
	bb.conn.Write([]byte("JOIN #" + bb.Channel + "\r\n"))

	rgb.YPrintf("[%s] Joined #%s as @%s!\n", timeStamp(), bb.Channel, bb.Name)
}

// Reads from the private credentials file and stores the data in the bot's Credentials field.
func (bb *TwitchBot) ReadCredentials() error {

	// reads from the file
	credFile, err := ioutil.ReadFile(bb.PrivatePath)
	if nil != err {
		return err
	}

	bb.Credentials = &OAuthCred{}

	// parses the file contents
	dec := json.NewDecoder(strings.NewReader(string(credFile)))
	if err = dec.Decode(bb.Credentials); nil != err && io.EOF != err {
		return err
	}

	return nil
}

// Makes the bot send a message to the chat channel.
func (bb *TwitchBot) Say(msg string) error {
	if "" == msg {
		return errors.New("msg was empty.")
	}
	_, err := bb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s %s\r\n", bb.Channel, msg)))
	if nil != err {
		return err
	}
	return nil
}

// Start
func (bb *TwitchBot) Start() {
	err := bb.ReadCredentials()
	db.ConnectToDb()
	if nil != err {
		fmt.Println(err)
		fmt.Println("Aborting...")
		return
	}

	for {
		bb.Connect()
		bb.JoinChannel()
		err = bb.HandleChat()
		if nil != err {
			time.Sleep(1000 * time.Millisecond)
			fmt.Println(err)
			fmt.Println("Starting bot again...")
		} else {
			return
		}
	}
}

func timeStamp() string {
	return TimeStamp(PSTFormat)
}

func TimeStamp(format string) string {
	return time.Now().Format(format)
}
