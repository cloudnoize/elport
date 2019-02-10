package pa

/*
#include <portaudio.h>
#cgo CFLAGS: -I/Users/elerer/portaudio/include
#cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework AudioUnit -framework CoreServices -framework Carbon -lstdc++ -L/Users/elerer/portaudio/lib/.libs/static -lportaudio
extern PaStreamCallback* paStreamCallback;
*/
import "C"

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
		fmt.Printf("PortAudio error: %s\n", C.Pa_GetErrorText(err))
	}
	return Error(err)
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
	return s, Error(err)
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
	Cbb(inputBuffer, outputBuffer unsafe.Pointer, frames uint64)
}

var Cba [1]IStream

//export streamCallback
func streamCallback(inputBuffer, outputBuffer unsafe.Pointer, frames C.ulong, timeInfo *C.PaStreamCallbackTimeInfo, statusFlags C.PaStreamCallbackFlags, userData unsafe.Pointer) {
	s := Cba[0]
	s.Cbb(inputBuffer, outputBuffer, uint64(frames))
}

func printType(args ...interface{}) {
	fmt.Println(reflect.TypeOf(args[0]))
}
