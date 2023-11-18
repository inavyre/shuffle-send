package main

import (
	"html/template"
	"log"
	"math/rand"
	"time"

	"github.com/alessioRaviola/shuffle-send/hub"
	"github.com/gin-gonic/gin"
)

var chars = [26]rune {
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
    'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

type RoomData struct {
    CreatedAt string
    ClientName string
    Token string
}

func main() {
    r := rand.New(rand.NewSource(time.Now().Unix()));

    // indexTemplate := template.Must(template.ParseFiles("templates/index.html"))
    // roomTemplate := template.Must(template.ParseFiles("templates/room.html"))

    newHub := hub.NewHub()
    go newHub.Run()
    log.Print("Started HUB")
    
    app := gin.Default()
    
    roomConnect := func (g *gin.Context) {
        token := g.Params.ByName("token")
        room := newHub.GetRoom(token)
        if (room == nil) {
            log.Printf("Tried to connect to %s but failed: no room", token)
            return
        }
        
        client, err := hub.ConnectClient(
            g.Writer, g.Request, g.Query("name"), room,
        )
        if (err != nil) {
            log.Printf("Error while connecting %s", err)
            return
        }

        log.Printf("New connected client to %s", token)

        go client.Read()
        go client.Write()
    }
    app.GET("/room/:token", roomConnect)

    room := func (g *gin.Context) {
        roomTemplate := template.Must(template.ParseFiles("templates/room.html"))

        name := g.Query("name")
        token := g.Query("token") 

        if token == "" {
            for i := 0; i < 4; i++ {
                token += string(chars[r.Intn(26)])
            }
            newHub.CreateRoom(token)
            log.Printf("Created room %s", token)
        } else {
            if !newHub.HasRoom(token) {
                log.Printf("Cannot join non existing room %s", token)
                return;
            }
            log.Printf("Joining room %s", token)
        }

        data := RoomData {
            CreatedAt: "",
            ClientName: name,
            Token: token,
        }
        templ := roomTemplate
        templ.Execute(g.Writer, data) 
    }
    app.GET("/room", room)

    index := func (g *gin.Context) {
        indexTemplate := template.Must(template.ParseFiles("templates/index.html"))
        templ := indexTemplate
        templ.Execute(g.Writer, nil)
    }
    app.GET("/", index)
    
    app.Static("/static", "./static")

    app.Run()
}
