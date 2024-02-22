package main

import (
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/interceptor/pkg/report"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

type memoizedFactory struct {
	i interceptor.Interceptor
}

func (m *memoizedFactory) NewInterceptor(_ string) (interceptor.Interceptor, error) {
	return m.i, nil
}

// Instead of creating a new factory for every PeerConnection this shares them
// across create calls
func newMemoizedFactory(f interceptor.Factory) interceptor.Factory {
	i, err := f.NewInterceptor("")
	if err != nil {
		panic(err)
	}

	return &memoizedFactory{i: i}
}

func configureNack(mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry) {
	generator, err := nack.NewGeneratorInterceptor()
	if err != nil {
		panic(err)
	}

	responder, err := nack.NewResponderInterceptor()
	if err != nil {
		panic(err)
	}

	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack"}, webrtc.RTPCodecTypeVideo)
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
	interceptorRegistry.Add(newMemoizedFactory(responder))
	interceptorRegistry.Add(newMemoizedFactory(generator))
}

func configureRTCPReports(interceptorRegistry *interceptor.Registry) {
	reciver, err := report.NewReceiverInterceptor()
	if err != nil {
		panic(err)
	}

	sender, err := report.NewSenderInterceptor()
	if err != nil {
		panic(err)
	}

	interceptorRegistry.Add(newMemoizedFactory(reciver))
	interceptorRegistry.Add(newMemoizedFactory(sender))
}

func configureTWCCHeaderExtensionSender(mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry) {
	if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	i, err := twcc.NewHeaderExtensionInterceptor()
	if err != nil {
		panic(err)
	}

	interceptorRegistry.Add(newMemoizedFactory(i))
}

func configureTWCCSender(mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry) {
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeVideo)
	if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeAudio)
	if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	generator, err := twcc.NewSenderInterceptor()
	if err != nil {
		panic(err)
	}

	interceptorRegistry.Add(newMemoizedFactory(generator))
}
