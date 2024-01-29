# pion-unbundled

`pion-unbundled` demonstrates how to use Pion with a unbundled remote

### Running

* `git clone https://github.com/sean-der/pion-unbundled.git`
* `cd pion-unbundled`
* `go run main.go`

In the command line you should see

```
Open http://localhost:8080 to access this demo
```

When opened you should see this on stdout

```
Audio PeerConnection ICEConnectionState(checking)
Video PeerConnection ICEConnectionState(checking)
Video PeerConnection ICEConnectionState(connected)
Audio PeerConnection ICEConnectionState(connected)
```

This means Chrome (one PeerConnection) is being connected against two PeerConnections in Go.
