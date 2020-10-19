package mediasoup

import (
	"encoding/json"
	"testing"

	"github.com/jiyeyuran/mediasoup-go/h264"
	"github.com/stretchr/testify/suite"
)

type ConsumerTestSuite struct {
	suite.Suite
	consumerDeviceCapabilities RtpCapabilities
	worker                     *Worker
	router                     *Router
	transport1                 ITransport
	transport2                 ITransport
	audioProducer              *Producer
	videoProducer              *Producer
}

func (suite *ConsumerTestSuite) SetupTest() {
	mediaCodecs := []RtpCodecCapability{
		{
			Kind:      "audio",
			MimeType:  "audio/opus",
			ClockRate: 48000,
			Channels:  2,
		},
		{
			Kind:      "video",
			MimeType:  "video/VP8",
			ClockRate: 90000,
		},
		{
			Kind:      "video",
			MimeType:  "video/H264",
			ClockRate: 90000,
			Parameters: RtpCodecSpecificParameters{
				RtpParameter: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     1,
					ProfileLevelId:        "4d0032",
				},
			},
		},
	}

	audioProducerParameters := ProducerOptions{
		Kind: MediaKind_Audio,
		RtpParameters: RtpParameters{
			Mid: "AUDIO",
			Codecs: []RtpCodecParameters{
				{
					MimeType:    "audio/opus",
					PayloadType: 111,
					ClockRate:   48000,
					Channels:    2,
					Parameters: RtpCodecSpecificParameters{
						Useinbandfec: 1,
						Usedtx:       1,
					},
				},
			},
			HeaderExtensions: []RtpHeaderExtensionParameters{
				{
					Uri: "urn:ietf:params:rtp-hdrext:sdes:mid",
					Id:  10,
				},
				{
					Uri: "urn:ietf:params:rtp-hdrext:ssrc-audio-level",
					Id:  12,
				},
			},
			Encodings: []RtpEncodingParameters{{Ssrc: 11111111}},
			Rtcp: RtcpParameters{
				Cname: "FOOBAR",
			},
		},
		AppData: H{"foo": 1, "bar": "2"},
	}

	videoProducerParameters := ProducerOptions{
		Kind: MediaKind_Video,
		RtpParameters: RtpParameters{
			Mid: "VIDEO",
			Codecs: []RtpCodecParameters{
				{
					MimeType:    "video/h264",
					PayloadType: 112,
					ClockRate:   90000,
					Parameters: RtpCodecSpecificParameters{
						RtpParameter: h264.RtpParameter{
							PacketizationMode: 1,
							ProfileLevelId:    "4d0032",
						},
					},
					RtcpFeedback: []RtcpFeedback{
						{Type: "nack", Parameter: ""},
						{Type: "nack", Parameter: "pli"},
						{Type: "goog-remb", Parameter: ""},
					},
				},
				{
					MimeType:    "video/rtx",
					PayloadType: 113,
					ClockRate:   90000,
					Parameters:  RtpCodecSpecificParameters{Apt: 112},
				},
			},
			HeaderExtensions: []RtpHeaderExtensionParameters{
				{
					Uri: "urn:ietf:params:rtp-hdrext:sdes:mid",
					Id:  10,
				},
				{
					Uri: "urn:3gpp:video-orientation",
					Id:  13,
				},
			},
			Encodings: []RtpEncodingParameters{
				{Ssrc: 22222222, Rtx: &RtpEncodingRtx{Ssrc: 22222223}},
				{Ssrc: 22222224, Rtx: &RtpEncodingRtx{Ssrc: 22222225}},
				{Ssrc: 22222226, Rtx: &RtpEncodingRtx{Ssrc: 22222227}},
				{Ssrc: 22222228, Rtx: &RtpEncodingRtx{Ssrc: 22222229}},
			},
			Rtcp: RtcpParameters{
				Cname: "FOOBAR",
			},
		},
		AppData: H{"foo": 1, "bar": "2"},
	}

	suite.consumerDeviceCapabilities = RtpCapabilities{
		Codecs: []RtpCodecCapability{
			{
				MimeType:             "audio/opus",
				Kind:                 "audio",
				PreferredPayloadType: 100,
				ClockRate:            48000,
				Channels:             2,
			},
			{
				MimeType:             "video/H264",
				Kind:                 "video",
				PreferredPayloadType: 101,
				ClockRate:            90000,
				Parameters: RtpCodecSpecificParameters{
					RtpParameter: h264.RtpParameter{
						LevelAsymmetryAllowed: 1,
						PacketizationMode:     1,
						ProfileLevelId:        "4d0032",
					},
				},
				RtcpFeedback: []RtcpFeedback{
					{Type: "nack", Parameter: ""},
					{Type: "nack", Parameter: "pli"},
					{Type: "ccm", Parameter: "fir"},
					{Type: "goog-remb", Parameter: ""},
				},
			},
			{
				MimeType:             "video/rtx",
				Kind:                 "video",
				PreferredPayloadType: 102,
				ClockRate:            90000,
				Parameters: RtpCodecSpecificParameters{
					Apt: 101,
				},
			},
		},
		HeaderExtensions: []RtpHeaderExtension{
			{
				Kind:             "audio",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:mid",
				PreferredId:      1,
				PreferredEncrypt: false,
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:mid",
				PreferredId:      1,
				PreferredEncrypt: false,
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
				PreferredId:      2,
				PreferredEncrypt: false,
			},
			{
				Kind:             "audio",
				Uri:              "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time", // eslint-disable-line max-len
				PreferredId:      4,
				PreferredEncrypt: false,
			},
			{
				Kind:             "video",
				Uri:              "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time", // eslint-disable-line max-len
				PreferredId:      4,
				PreferredEncrypt: false,
			},
			{
				Kind:             "audio",
				Uri:              "urn:ietf:params:rtp-hdrext:ssrc-audio-level",
				PreferredId:      10,
				PreferredEncrypt: false,
			},
			{
				Kind:             "video",
				Uri:              "urn:3gpp:video-orientation",
				PreferredId:      11,
				PreferredEncrypt: false,
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:toffset",
				PreferredId:      12,
				PreferredEncrypt: false,
			},
		},
	}

	suite.worker = CreateTestWorker()
	suite.router, _ = suite.worker.CreateRouter(RouterOptions{MediaCodecs: mediaCodecs})

	suite.transport1, _ = suite.router.CreateWebRtcTransport(WebRtcTransportOptions{
		ListenIps: []TransportListenIp{
			{Ip: "127.0.0.1"},
		},
	})
	suite.transport2, _ = suite.router.CreateWebRtcTransport(WebRtcTransportOptions{
		ListenIps: []TransportListenIp{
			{Ip: "127.0.0.1"},
		},
	})

	suite.audioProducer, _ = suite.transport1.Produce(audioProducerParameters)
	suite.videoProducer, _ = suite.transport1.Produce(videoProducerParameters)

	// Pause the videoProducer.
	suite.videoProducer.Pause()
}

func (suite *ConsumerTestSuite) TearDownTest() {
	suite.worker.Close()
}

func (suite *ConsumerTestSuite) TestTransportConsume_Succeeds() {
	router, transport2 := suite.router, suite.transport2

	onObserverNewConsumer1 := NewMockFunc(suite.T())

	transport2.Observer().Once("newconsumer", onObserverNewConsumer1.Fn())

	suite.True(router.CanConsume(suite.audioProducer.Id(), suite.consumerDeviceCapabilities))

	audioConsumer, err := transport2.Consume(ConsumerOptions{
		ProducerId:      suite.audioProducer.Id(),
		RtpCapabilities: suite.consumerDeviceCapabilities,
		AppData:         H{"baz": "LOL"},
	})

	onObserverNewConsumer1.Wait()

	suite.NoError(err)
	onObserverNewConsumer1.ExpectCalledTimes(1)
	onObserverNewConsumer1.ExpectCalledWith(audioConsumer)
	suite.NotEmpty(audioConsumer.Id())
	suite.Equal(suite.audioProducer.Id(), audioConsumer.ProducerId())
	suite.False(audioConsumer.Closed())
	suite.EqualValues(MediaKind_Audio, audioConsumer.Kind())
	suite.NotEmpty(audioConsumer.RtpParameters())
	suite.Equal("0", audioConsumer.RtpParameters().Mid)
	suite.Len(audioConsumer.RtpParameters().Codecs, 1)

	data, _ := json.Marshal(audioConsumer.RtpParameters().Codecs[0])
	suite.JSONEq(`
	{
		"mimeType" : "audio/opus",
		"clockRate" : 48000,
		"payloadType" : 100,
		"channels" : 2,
		"parameters" : {
		  "useinbandfec" : 1,
		  "usedtx" : 1
		}
	  }
	`, string(data))

	suite.Equal(ConsumerType_Simple, audioConsumer.Type())
	suite.False(audioConsumer.Paused())
	suite.False(audioConsumer.ProducerPaused())
	suite.Equal(ConsumerScore{Score: 10, ProducerScore: 0, ProducerScores: []uint32{0}}, audioConsumer.Score())
	suite.Nil(audioConsumer.CurrentLayers())
	suite.Nil(audioConsumer.PreferredLayers())
	suite.Equal(H{"baz": "LOL"}, audioConsumer.AppData())

	routerDump, _ := router.Dump()

	suite.Equal([]string{audioConsumer.Id()}, routerDump.MapProducerIdConsumerIds[suite.audioProducer.Id()])
	suite.Equal(suite.audioProducer.Id(), routerDump.MapConsumerIdProducerId[audioConsumer.Id()])

	transportDump, _ := transport2.Dump()

	suite.Equal(transport2.Id(), transportDump.Id)
	suite.Equal([]string{audioConsumer.Id()}, transportDump.ConsumerIds)

	onObserverNewConsumer2 := NewMockFunc(suite.T())
	transport2.Observer().Once("newconsumer", onObserverNewConsumer2.Fn())

	suite.True(router.CanConsume(suite.videoProducer.Id(), suite.consumerDeviceCapabilities))

	videoConsumer, err := transport2.Consume(ConsumerOptions{
		ProducerId:      suite.videoProducer.Id(),
		RtpCapabilities: suite.consumerDeviceCapabilities,
		Paused:          true,
		PreferredLayers: &ConsumerLayers{SpatialLayer: 12},
		AppData:         H{"baz": "LOL"},
	})
	suite.NoError(err)

	onObserverNewConsumer2.Wait()

	onObserverNewConsumer2.ExpectCalledTimes(1)
	onObserverNewConsumer2.ExpectCalledWith(videoConsumer)
	suite.NotEmpty(videoConsumer.Id())
	suite.Equal(suite.videoProducer.Id(), videoConsumer.ProducerId())
	suite.False(videoConsumer.Closed())
	suite.EqualValues(MediaKind_Video, videoConsumer.Kind())
	suite.Equal("1", videoConsumer.RtpParameters().Mid)
	suite.Len(videoConsumer.RtpParameters().Codecs, 2)
	data, _ = json.Marshal(videoConsumer.RtpParameters().Codecs[0])
	suite.JSONEq(`
	{
		"mimeType" : "video/H264",
		"clockRate" : 90000,
		"payloadType" : 103,
		"parameters" : {
		  "packetization-mode" : 1,
		  "profile-level-id" : "4d0032"
		},
		"rtcpFeedback" : [
		  { "type":"nack" },
		  {
			"type" : "nack",
			"parameter" : "pli"
		  },
		  { "type":"ccm", "parameter":"fir" },
		  { "type":"goog-remb" }
		]
	  }
	`, string(data))
	data, _ = json.Marshal(videoConsumer.RtpParameters().Codecs[1])
	suite.JSONEq(`
	{
		"mimeType" : "video/rtx",
		"clockRate" : 90000,
		"payloadType" : 104,
		"parameters" : { "apt":103 }
	  }
	`, string(data))
	suite.EqualValues(ConsumerType_Simulcast, videoConsumer.Type())
	suite.True(videoConsumer.Paused())
	suite.True(videoConsumer.ProducerPaused())
	suite.Equal(ConsumerScore{Score: 10, ProducerScore: 0, ProducerScores: []uint32{0, 0, 0, 0}}, videoConsumer.Score())
	suite.Nil(videoConsumer.CurrentLayers())
	suite.Equal(H{"baz": "LOL"}, videoConsumer.AppData())

	routerDump, _ = router.Dump()

	expectedRouterDump := RouterDump{
		MapProducerIdConsumerIds: map[string][]string{
			suite.audioProducer.Id(): {audioConsumer.Id()},
			suite.videoProducer.Id(): {videoConsumer.Id()},
		},
		MapConsumerIdProducerId: map[string]string{
			audioConsumer.Id(): suite.audioProducer.Id(),
			videoConsumer.Id(): suite.videoProducer.Id(),
		},
	}

	suite.Equal(expectedRouterDump.MapProducerIdConsumerIds, routerDump.MapProducerIdConsumerIds)
	suite.Equal(expectedRouterDump.MapConsumerIdProducerId, routerDump.MapConsumerIdProducerId)

	transportDump, _ = transport2.Dump()
	expectedTransportDump := TransportDump{
		Id:          transport2.Id(),
		ConsumerIds: []string{videoConsumer.Id(), audioConsumer.Id()},
	}

	suite.Equal(expectedTransportDump.Id, transportDump.Id)
	suite.Equal(expectedTransportDump.ConsumerIds, transportDump.ConsumerIds)
}

func (suite *ConsumerTestSuite) TestTransportConsume_UnsupportedError() {
	router, transport2, audioProducer := suite.router, suite.transport2, suite.audioProducer
	invalidDeviceCapabilitiesJSON := `
	{
		"codecs" : [
		  {
			"kind" : "audio",
			"mimeType" : "audio/ISAC",
			"clockRate" : 32000,
			"preferredPayloadType" : 100,
			"channels" : 1
		  }
		],
		"headerExtensions" : []
	  }
	`
	var invalidDeviceCapabilities RtpCapabilities

	json.Unmarshal([]byte(invalidDeviceCapabilitiesJSON), &invalidDeviceCapabilities)

	suite.False(router.CanConsume(audioProducer.Id(), invalidDeviceCapabilities))

	_, err := transport2.Consume(ConsumerOptions{
		ProducerId:      audioProducer.Id(),
		RtpCapabilities: invalidDeviceCapabilities,
	})
	suite.IsType(NewUnsupportedError(""), err)

	invalidDeviceCapabilities = RtpCapabilities{}

	suite.False(router.CanConsume(audioProducer.Id(), invalidDeviceCapabilities))

	_, err = transport2.Consume(ConsumerOptions{
		ProducerId:      audioProducer.Id(),
		RtpCapabilities: invalidDeviceCapabilities,
	})
	suite.IsType(NewUnsupportedError(""), err)
}

func (suite *ConsumerTestSuite) TestConsumerDump() {
	type Dump struct {
		RtpParameters              RtpParameters
		Id                         string
		Kind                       string
		Type                       string
		ConsumableRtpEncodings     []RtpMappingEncoding
		SupportedCodecPayloadTypes []uint32
		Paused                     bool
		ProducerPaused             bool
	}
	var data Dump
	audioConsumer := suite.audioConsumer()
	audioConsumer.Dump().Unmarshal(&data)

	suite.Equal(audioConsumer.Id(), data.Id)
	suite.EqualValues(audioConsumer.Kind(), data.Kind)
	suite.NotEmpty(data.RtpParameters)
	suite.Len(data.RtpParameters.Codecs, 1)
	suite.Equal("audio/opus", data.RtpParameters.Codecs[0].MimeType)
	suite.EqualValues(100, data.RtpParameters.Codecs[0].PayloadType)
	suite.EqualValues(48000, data.RtpParameters.Codecs[0].ClockRate)
	suite.EqualValues(2, data.RtpParameters.Codecs[0].Channels)
	suite.Equal(RtpCodecSpecificParameters{Useinbandfec: 1, Usedtx: 1}, data.RtpParameters.Codecs[0].Parameters)
	suite.Equal([]RtcpFeedback{}, data.RtpParameters.Codecs[0].RtcpFeedback)
	suite.Len(data.RtpParameters.HeaderExtensions, 3)
	headerExtensionsJSON, _ := json.Marshal(data.RtpParameters.HeaderExtensions)
	suite.JSONEq(`
	[
		{
		  "uri" : "urn:ietf:params:rtp-hdrext:ssrc-audio-level",
		  "id" : 1,
		  "parameters" : {},
		  "encrypt"    : false
		},
		{
		  "uri" : "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
		  "id" : 3,
		  "parameters" : {},
		  "encrypt"    : false
		}
	  ]
	`, string(headerExtensionsJSON))
	suite.Len(data.RtpParameters.Encodings, 1)
	suite.Equal([]RtpEncodingParameters{
		{CodecPayloadType: 100, Ssrc: audioConsumer.RtpParameters().Encodings[0].Ssrc},
	}, data.RtpParameters.Encodings)
	suite.EqualValues("simple", data.Type)
	suite.Len(data.ConsumableRtpEncodings, 1)
	suite.Equal([]RtpMappingEncoding{
		{Ssrc: suite.audioProducer.ConsumableRtpParameters().Encodings[0].Ssrc},
	}, data.ConsumableRtpEncodings)
	suite.Equal([]uint32{100}, data.SupportedCodecPayloadTypes)
	suite.False(data.Paused)
	suite.False(data.ProducerPaused)

	videoConsumer := suite.videoConsumer(true)
	data = Dump{}
	videoConsumer.Dump().Unmarshal(&data)

	suite.Equal(videoConsumer.Id(), data.Id)
	suite.EqualValues(videoConsumer.Kind(), data.Kind)
	suite.NotEmpty(data.RtpParameters)
	suite.Len(data.RtpParameters.Codecs, 2)
	suite.Equal("video/H264", data.RtpParameters.Codecs[0].MimeType)
	suite.EqualValues(103, data.RtpParameters.Codecs[0].PayloadType)
	suite.Equal(90000, data.RtpParameters.Codecs[0].ClockRate)
	suite.Empty(data.RtpParameters.Codecs[0].Channels)
	suite.Equal(h264.RtpParameter{
		PacketizationMode: 1,
		ProfileLevelId:    "4d0032",
	}, data.RtpParameters.Codecs[0].Parameters.RtpParameter)
	suite.EqualValues([]RtcpFeedback{
		{Type: "nack"},
		{Type: "nack", Parameter: "pli"},
		{Type: "ccm", Parameter: "fir"},
		{Type: "goog-remb"},
	}, data.RtpParameters.Codecs[0].RtcpFeedback)
	suite.Len(data.RtpParameters.HeaderExtensions, 4)
	headerExtensionsJSON, _ = json.Marshal(data.RtpParameters.HeaderExtensions)
	suite.JSONEq(`
	[
		{
		  "uri" : "urn:ietf:params:rtp-hdrext:toffset",
		  "id" : 2,
		  "parameters" : {},
		  "encrypt" : false
		},
		{
		  "uri" : "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
		  "id" : 3,
		  "parameters" : {},
		  "encrypt" : false
		},
		{
		  "uri" : "urn:3gpp:video-orientation",
		  "id" : 4,
		  "parameters" : {},
		  "encrypt" : false
		}
	  ]
	`, string(headerExtensionsJSON))
	suite.Len(data.RtpParameters.Encodings, 1)
	suite.EqualValues([]RtpEncodingParameters{
		{
			CodecPayloadType: 103,
			Ssrc:             videoConsumer.RtpParameters().Encodings[0].Ssrc,
			Rtx: &RtpEncodingRtx{
				Ssrc: videoConsumer.RtpParameters().Encodings[0].Rtx.Ssrc,
			},
		},
	}, data.RtpParameters.Encodings)
	suite.Len(data.ConsumableRtpEncodings, 4)
	suite.EqualValues([]RtpMappingEncoding{
		{Ssrc: suite.videoProducer.ConsumableRtpParameters().Encodings[0].Ssrc},
		{Ssrc: suite.videoProducer.ConsumableRtpParameters().Encodings[1].Ssrc},
		{Ssrc: suite.videoProducer.ConsumableRtpParameters().Encodings[2].Ssrc},
		{Ssrc: suite.videoProducer.ConsumableRtpParameters().Encodings[3].Ssrc},
	}, data.ConsumableRtpEncodings)
	suite.Equal([]uint32{103}, data.SupportedCodecPayloadTypes)
	suite.True(data.Paused)
	suite.True(data.ProducerPaused)
}

func (suite *ConsumerTestSuite) TestConsumerGetStats() {
	audioConsumer := suite.audioConsumer()

	consumerStats, err := audioConsumer.GetStats()
	suite.NoError(err)

	expectedConsumerStats := []ConsumerStat{}
	for _, stats := range consumerStats {
		expectedConsumerStats = append(expectedConsumerStats, ConsumerStat{
			Type:     stats.Type,
			Kind:     stats.Kind,
			MimeType: stats.MimeType,
			Ssrc:     stats.Ssrc,
		})
	}

	suite.Contains(expectedConsumerStats, ConsumerStat{
		Type:     "outbound-rtp",
		Kind:     "audio",
		MimeType: "audio/opus",
		Ssrc:     audioConsumer.RtpParameters().Encodings[0].Ssrc,
	})

	videoConsumer := suite.videoConsumer(false)
	consumerStats, err = videoConsumer.GetStats()
	suite.NoError(err)

	expectedConsumerStats = []ConsumerStat{}
	for _, stats := range consumerStats {
		expectedConsumerStats = append(expectedConsumerStats, ConsumerStat{
			Type:     stats.Type,
			Kind:     stats.Kind,
			MimeType: stats.MimeType,
			Ssrc:     stats.Ssrc,
		})
	}

	suite.Contains(expectedConsumerStats, ConsumerStat{
		Type:     "outbound-rtp",
		Kind:     "video",
		MimeType: "video/H264",
		Ssrc:     videoConsumer.RtpParameters().Encodings[0].Ssrc,
	})
}

func (suite *ConsumerTestSuite) TestConsumerPauseAndResume() {
	audioConsumer := suite.audioConsumer()

	audioConsumer.Pause()

	suite.True(audioConsumer.Paused())
	var data struct{ Paused bool }
	audioConsumer.Dump().Unmarshal(&data)
	suite.True(data.Paused)

	audioConsumer.Resume()

	suite.False(audioConsumer.Paused())
	audioConsumer.Dump().Unmarshal(&data)
	suite.False(data.Paused)
}

func (suite *ConsumerTestSuite) TestConsumerEmitsProducerpauseAndProducerresume() {
	audioConsumer := suite.audioConsumer()

	wf := NewMockFunc(suite.T())

	audioConsumer.On("producerpause", wf.Fn())
	suite.audioProducer.Pause()

	wf.Wait()

	suite.False(audioConsumer.Paused())
	suite.True(audioConsumer.ProducerPaused())

	audioConsumer.On("producerresume", wf.Fn())
	suite.audioProducer.Resume()

	wf.Wait()

	suite.False(audioConsumer.Paused())
	suite.False(audioConsumer.ProducerPaused())
}

func (suite *ConsumerTestSuite) TestConsumerEmitsScore() {
	audioConsumer := suite.audioConsumer()

	onScore := NewMockFunc(suite.T())
	audioConsumer.On("score", onScore.Fn())

	channel := audioConsumer.channel

	channel.Emit(audioConsumer.Id(), "score", json.RawMessage(`{"producerScore": 10, "score": 9}`))
	channel.Emit(audioConsumer.Id(), "score", json.RawMessage(`{"producerScore": 9, "score": 9}`))
	channel.Emit(audioConsumer.Id(), "score", json.RawMessage(`{"producerScore": 8, "score": 8}`))

	onScore.Wait()
	onScore.ExpectCalledTimes(3)
	suite.Equal(ConsumerScore{ProducerScore: 8, Score: 8}, audioConsumer.Score())
}

func (suite *ConsumerTestSuite) TestConsumerClose() {
	audioConsumer := suite.audioConsumer()
	videoConsumer := suite.videoConsumer(true)

	onObserverClose := NewMockFunc(suite.T())

	audioConsumer.Observer().Once("close", onObserverClose.Fn())
	audioConsumer.Close()

	onObserverClose.Wait()
	onObserverClose.ExpectCalledTimes(1)
	suite.True(audioConsumer.Closed())

	routerDump, _ := suite.router.Dump()

	suite.Empty(routerDump.MapProducerIdConsumerIds[suite.audioProducer.Id()])
	suite.Equal(suite.videoProducer.Id(), routerDump.MapConsumerIdProducerId[videoConsumer.Id()])

	transportDump, _ := suite.transport2.Dump()

	suite.Equal(suite.transport2.Id(), transportDump.Id)
	suite.Empty(transportDump.ProducerIds)
	suite.EqualValues([]string{videoConsumer.Id()}, transportDump.ConsumerIds)
}

func (suite *ConsumerTestSuite) TestConsumerRejectIfClosed() {
	audioConsumer := suite.audioConsumer()
	audioConsumer.Close()

	suite.Error(audioConsumer.Dump().Err)

	_, err := audioConsumer.GetStats()

	suite.Error(err)
	suite.Error(audioConsumer.Pause())
	suite.Error(audioConsumer.Resume())
	suite.Error(audioConsumer.SetPreferredLayers(ConsumerLayers{}))
	suite.Error(audioConsumer.RequestKeyFrame())
}

func (suite *ConsumerTestSuite) TestConsumerEmitsProducerClosed() {
	audioConsumer := suite.audioConsumer()

	onObserverClose := NewMockFunc(suite.T())
	audioConsumer.Observer().Once("close", onObserverClose.Fn())

	wf := NewMockFunc(suite.T())

	audioConsumer.On("producerclose", wf.Fn())

	suite.audioProducer.Close()

	wf.Wait()

	onObserverClose.Wait()
	onObserverClose.ExpectCalledTimes(1)
	suite.True(audioConsumer.Closed())
}

func (suite *ConsumerTestSuite) TestConsumerEmitsTransportClosed() {
	videoConsumer := suite.videoConsumer(false)

	onObserverClose := NewMockFunc(suite.T())
	videoConsumer.Observer().Once("close", onObserverClose.Fn())

	wf := NewMockFunc(suite.T())

	videoConsumer.Observer().On("transportclose", wf.Fn())

	suite.transport2.Close()

	onObserverClose.Wait()
	onObserverClose.ExpectCalledTimes(1)
	suite.True(videoConsumer.Closed())

	routerDump, _ := suite.router.Dump()

	suite.Empty(routerDump.MapProducerIdConsumerIds[suite.audioProducer.Id()])
	suite.Empty(routerDump.MapConsumerIdProducerId)
}

func (suite *ConsumerTestSuite) audioConsumer() *Consumer {
	audioConsumer, err := suite.transport2.Consume(ConsumerOptions{
		ProducerId:      suite.audioProducer.Id(),
		RtpCapabilities: suite.consumerDeviceCapabilities,
		AppData:         H{"baz": "LOL"},
	})
	suite.Require().NoError(err)

	return audioConsumer
}

func (suite *ConsumerTestSuite) videoConsumer(paused bool) *Consumer {
	videoConsumer, _ := suite.transport2.Consume(ConsumerOptions{
		ProducerId:      suite.videoProducer.Id(),
		RtpCapabilities: suite.consumerDeviceCapabilities,
		Paused:          paused,
		AppData:         H{"baz": "LOL"},
	})

	return videoConsumer
}

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}