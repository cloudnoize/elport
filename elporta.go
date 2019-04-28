package pa

/*
#include <portaudio.h>
#cgo CFLAGS: -I/Users/elerer/portaudio/include
#cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework AudioUnit -framework CoreServices -framework Carbon -lstdc++ -L/Users/elerer/portaudio/lib/.libs/static -lportaudio
extern PaStreamCallback* paStreamCallback;
*/
import "C"

//Replace Path to your Path to the PortAudio include dir and Libs

import (
	"fmt"
	"reflect"
	"unsafe"
)

type SampleFormat uint64

const (
	Float32 SampleFormat = 0x00000001
	Int32   SampleFormat = 0x00000002
	Int24   SampleFormat = 0x00000004
	Int16   SampleFormat = 0x00000008
	Int8    SampleFormat = 0x00000010
	UInt8   SampleFormat = 0x00000020
)

// Error wraps over PaError.
type Error C.PaError

func (err Error) Error() string {
	return C.GoString(C.Pa_GetErrorText(C.PaError(err)))
}

// VersionText returns the textual description of the PortAudio release.
func VersionText() string {
	return C.GoString(C.Pa_GetVersionText())
}

func Initialize() error {
	err := C.Pa_Initialize()
	if err != C.paNoError {
		return Error(err)
	}
	return nil

}

func ListDevices() error {
	numDevices := C.Pa_GetDeviceCount()
	if numDevices < 0 {
		fmt.Printf("ERROR: Pa_CountDevices returned 0x%x\n", numDevices)
		return Error(C.paInvalidDevice)
	}

	dis := make([]*C.PaDeviceInfo, numDevices)

	for i := 0; i < int(numDevices); i++ {
		//x := C.PaDeviceIndex(i)
		dis[i] = C.Pa_GetDeviceInfo(C.int(i))
	}

	for n, di := range dis {
		nm := C.GoString(di.name)
		fmt.Printf("device [%d]: name [%s], inputs [%d], outputs [%d], default sample rate [%f]\n", n, nm, di.maxInputChannels, di.maxOutputChannels, di.defaultSampleRate)
	}
	return nil
}

func Terminate() error {
	err := C.Pa_Terminate()
	if err != C.paNoError {
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
	}
	return Error(err)

}

func OpenDefaultStream(numIn, numOut int, sf SampleFormat, sampleRate float64, framesPerBuffer uint64, is IStream) (*Stream, error) {
	s := &Stream{}
	err := C.Pa_OpenDefaultStream(&s.stream, C.int(numIn), C.int(numOut), C.PaSampleFormat(sf), C.double(sampleRate), C.ulong(framesPerBuffer), C.paStreamCallback, unsafe.Pointer(&is))
	if err != C.paNoError {
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
		return nil, Error(err)
	}
	return s, nil
}

func OpenStream(in, out *PaStreamParameters, sf SampleFormat, sampleRate uint64, framesPerBuffer uint64) (*Stream, error) {
	s := &Stream{}
	pa_in := &C.PaStreamParameters{}
	if in != nil {
		pa_in.channelCount = C.int(in.ChannelCount)
		pa_in.device = C.int(in.DeviceNum)
		pa_in.sampleFormat = C.ulong(in.Sampleformat)
		pa_in.suggestedLatency = C.Pa_GetDeviceInfo(pa_in.device).defaultLowInputLatency
		pa_in.hostApiSpecificStreamInfo = nil
	} else {
		pa_in = nil
	}

	pa_out := &C.PaStreamParameters{}
	if out != nil {
		pa_out.channelCount = C.int(out.ChannelCount)
		pa_out.device = C.int(out.DeviceNum)
		pa_out.sampleFormat = C.ulong(out.Sampleformat)
		pa_out.suggestedLatency = C.Pa_GetDeviceInfo(pa_out.device).defaultLowInputLatency
		pa_out.hostApiSpecificStreamInfo = nil
	} else {
		pa_out = nil
	}
	err := C.Pa_OpenStream(&s.stream, pa_in, pa_out, C.double(sampleRate), C.ulong(framesPerBuffer), C.paClipOff, C.paStreamCallback, nil)
	if err != C.paNoError {
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
		return nil, Error(err)
	}
	return s, nil
}

type Stream struct {
	stream unsafe.Pointer
}

func (s *Stream) Start() error {
	err := C.Pa_StartStream(s.stream)
	if err != C.paNoError {
		fmt.Println("PortAudio error: ", C.GoString(C.Pa_GetErrorText(err)))
	}
	return Error(err)
}

func (s *Stream) Close() error {
	fmt.Println("Closing stream")
	err := C.Pa_CloseStream(unsafe.Pointer(s.stream))
	if err != C.paNoError {
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
	}
	return Error(err)
}

func (s *Stream) Stop() error {
	err := C.Pa_StopStream(unsafe.Pointer(s.stream))
	if err != C.paNoError {
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
	}
	return Error(err)
}

type IStream interface {
	CallBack(inputBuffer, outputBuffer unsafe.Pointer, frames uint64)
}

var CbStream IStream

//export streamCallback
func streamCallback(inputBuffer, outputBuffer unsafe.Pointer, frames C.ulong, timeInfo *C.PaStreamCallbackTimeInfo, statusFlags C.PaStreamCallbackFlags, userData unsafe.Pointer) {
	CbStream.CallBack(inputBuffer, outputBuffer, uint64(frames))
}

func printType(args ...interface{}) {
	fmt.Println(reflect.TypeOf(args[0]))
}

type PaStreamParameters struct {
	DeviceNum    int
	ChannelCount int
	Sampleformat SampleFormat
}

func (p *PaStreamParameters) String() string {
	return fmt.Sprintf("device num %d, channel count %d, sampleformat %d ", p.DeviceNum, p.ChannelCount, p.Sampleformat)
}

func IsformatSupported(in, out *PaStreamParameters, desiredSampleRate uint64) error {
	pa_in := &C.PaStreamParameters{}
	if in != nil {
		pa_in.channelCount = C.int(in.ChannelCount)
		pa_in.device = C.int(in.DeviceNum)
		pa_in.sampleFormat = C.ulong(in.Sampleformat)
		pa_in.suggestedLatency = C.Pa_GetDeviceInfo(pa_in.device).defaultLowInputLatency
		pa_in.hostApiSpecificStreamInfo = nil
	} else {
		pa_in = nil
	}

	pa_out := &C.PaStreamParameters{}
	if out != nil {
		pa_out.channelCount = C.int(out.ChannelCount)
		pa_out.device = C.int(out.DeviceNum)
		pa_out.sampleFormat = C.ulong(out.Sampleformat)
		pa_out.suggestedLatency = C.Pa_GetDeviceInfo(pa_out.device).defaultLowInputLatency
		pa_out.hostApiSpecificStreamInfo = nil
	} else {
		pa_out = nil
	}

	err := C.Pa_IsFormatSupported(pa_in, pa_out, C.double(desiredSampleRate))

	if err != C.paFormatIsSupported {
		return Error(err)
	}

	return nil

}
