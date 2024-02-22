package main

import (
	"fmt"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

type NoBundlePeerConnection struct {
	audioPeerConnection *webrtc.PeerConnection
	videoPeerConnection *webrtc.PeerConnection
}

func NewNoBundlePeerConnection() *NoBundlePeerConnection {
	i := &interceptor.Registry{}

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}

	configureNack(m, i)
	configureRTCPReports(i)
	configureTWCCHeaderExtensionSender(m, i)
	configureTWCCSender(m, i)

	api := webrtc.NewAPI(webrtc.WithInterceptorRegistry(i), webrtc.WithMediaEngine(m))

	audioPeerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	if _, err = audioPeerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	videoPeerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	if _, err = videoPeerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	audioPeerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		fmt.Printf("Audio PeerConnection ICEConnectionState(%s) \n", i.String())
	})

	videoPeerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		fmt.Printf("Video PeerConnection ICEConnectionState(%s) \n", i.String())
	})

	audioPeerConnection.OnTrack(onTrackHandler)
	videoPeerConnection.OnTrack(onTrackHandler)

	return &NoBundlePeerConnection{
		audioPeerConnection: audioPeerConnection,
		videoPeerConnection: videoPeerConnection,
	}
}

func (n *NoBundlePeerConnection) CreateOffer() []byte {
	audioOffer := getGatheredOffer(n.audioPeerConnection)
	videoOffer := getGatheredOffer(n.videoPeerConnection)

	audioOffer.MediaDescriptions[0].Attributes = append(getCertificateFingerprint(audioOffer.Attributes), audioOffer.MediaDescriptions[0].Attributes...)
	videoOffer.MediaDescriptions[0].Attributes = append(getCertificateFingerprint(videoOffer.Attributes), videoOffer.MediaDescriptions[0].Attributes...)

	for i := range videoOffer.MediaDescriptions[0].Attributes {
		if videoOffer.MediaDescriptions[0].Attributes[i].Key == "mid" {
			videoOffer.MediaDescriptions[0].Attributes[i].Value = "1"
		}
	}

	unbundledOffer := sdp.SessionDescription{
		Version:          audioOffer.Version,
		Origin:           audioOffer.Origin,
		SessionName:      audioOffer.SessionName,
		TimeDescriptions: audioOffer.TimeDescriptions,
		MediaDescriptions: []*sdp.MediaDescription{
			audioOffer.MediaDescriptions[0],
			videoOffer.MediaDescriptions[0],
		},
	}

	marshaled, err := unbundledOffer.Marshal()
	if err != nil {
		panic(err)
	}

	return marshaled
}

func (n *NoBundlePeerConnection) SetRemoteDescription(answer []byte) {
	parsed := sdp.SessionDescription{}
	if err := parsed.Unmarshal(answer); err != nil {
		panic(err)
	}

	mediaDescriptions := append([]*sdp.MediaDescription{}, parsed.MediaDescriptions...)
	parsed.MediaDescriptions = nil

	parsed.MediaDescriptions = []*sdp.MediaDescription{mediaDescriptions[0]}
	marshaled, err := parsed.Marshal()
	if err != nil {
		panic(err)
	} else if err := n.audioPeerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: string(marshaled)}); err != nil {
		panic(err)
	}

	parsed.MediaDescriptions = []*sdp.MediaDescription{mediaDescriptions[1]}
	marshaled, err = parsed.Marshal()
	if err != nil {
		panic(err)
	} else if err := n.videoPeerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: strings.Replace(string(marshaled), "mid:1", "mid:0", -1)}); err != nil {
		panic(err)
	}
}

func (n *NoBundlePeerConnection) Close() {
	if err := n.audioPeerConnection.Close(); err != nil {
		panic(err)
	}
	if err := n.videoPeerConnection.Close(); err != nil {
		panic(err)
	}
}

func onTrackHandler(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	for {
		if _, _, err := t.ReadRTP(); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func getGatheredOffer(peerConnection *webrtc.PeerConnection) sdp.SessionDescription {
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	<-gatherComplete

	parsed := sdp.SessionDescription{}
	if err = parsed.Unmarshal([]byte(peerConnection.LocalDescription().SDP)); err != nil {
		panic(err)
	}

	return parsed
}

func getCertificateFingerprint(attributes []sdp.Attribute) []sdp.Attribute {
	for _, a := range attributes {
		if a.Key == "fingerprint" {
			return []sdp.Attribute{a}
		}
	}

	panic("fingerprint attribute not found")
}
