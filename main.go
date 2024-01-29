package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

const indexHtml = `
<html>
  <head>
    <title>pion-unbundled</title>
  </head>

  <body>
    <video style="width: 1280" controls autoplay id="video"> </video>
  </body>

  <script>
	const peerConnection = new RTCPeerConnection()

    navigator.mediaDevices.getUserMedia({audio: true, video: true}).then(stream => {
      stream.getTracks().forEach(t => peerConnection.addTrack(t))

  	  const wsUri = ` + "`" + `ws://${window.location.host}/websocket` + "`" + `

	  const ws = new WebSocket(wsUri);

	  peerConnection.onicecandidate = evt => {
	  	if (evt.candidate == null) {
		  ws.send(peerConnection.localDescription.sdp)
		}
	  }

	  ws.onopen = evt => {
	  	console.log("websocket open");
	  }
	  ws.onclose = evt => {
	  	console.log("websocket close");
	  }
	  ws.onmessage = evt => {
		peerConnection.setRemoteDescription({type: 'offer', sdp: evt.data})
		peerConnection.createAnswer().then(answer => {
			peerConnection.setLocalDescription(answer)
		})
	  }
	  ws.onerror = evt => {
	  	console.log("websocket error: " + evt.data);
	  }

    })
  </script>
</html>
`

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	p := NewNoBundlePeerConnection()
	offer := p.CreateOffer()

	if err = c.WriteMessage(websocket.TextMessage, offer); err != nil {
		p.Close()
		panic(err)
	}

	if _, message, err := c.ReadMessage(); err != nil {
		p.Close()
		panic(err)
	} else {
		p.SetRemoteDescription(message)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, indexHtml)
	})
	http.HandleFunc("/websocket", echo)

	fmt.Println("Listening on :8080")
	panic(http.ListenAndServe(":8080", nil))
}
