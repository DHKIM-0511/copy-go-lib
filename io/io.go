// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// io 패키지는 I/O(입출력) 프리미티브(원시 기능)를 위한 기본 인터페이스들을 제공합니다.
// 이 패키지의 주된 역할은 os 패키지 등에 존재하는 기존의 I/O 구현체들을
// 기능을 추상화한 '공유 공용 인터페이스(shared public interfaces)'로 래핑(감싸기)하고,
// 그 외 관련된 몇 가지 프리미티브를 추가로 제공하는 것입니다.

// 이러한 인터페이스와 프리미티브는 다양한 구현을 가진 저수준 연산들을 래핑하고 있으므로,
// 별도로 명시되지 않는 한, 클라이언트(사용자)는 이것들이 병렬 실행에 안전하다(Thread-safe하다)고 가정해서는 안 됩니다.

package io

import (
	"errors"
)

/*
var vs const
SeekStart같은 변수는 불변인 const로 되어있음.
EOF나 ErrShortWrite는 var로 정의됨 -> errors.New가 런타임에 객체(포인터)를 생성하기 때문
*/

// Seek whence values
// Seek(탐색) 작업의 기준 값(whence)들입니다
const (
	SeekStart   = 0
	SeekCurrent = 1
	SeekEnd     = 2
)

/*
ErrShortWrite(파스칼케이스): 외부 패키지에서 사용
errInvalidWrite(카멜케이스): io 패키지 내부에서만 사용하는 private 에러.
 구현체가 엉뚱한 값을 리턴하는경우 내부적으로 체크하는 용도
*/
// ErrShortWrite는 쓰기 작업이 요청된 것보다 적은 바이트를처리했지만,
// 명시적인 오류를 반환하지는 않는 경우를 의미합니다
var ErrShortWrite = errors.New("short write")

// errInvalidWrite는 쓰기 작업이 불가능한 개수(음수이거나 요청보다 큰 값)을 반환했음을 의미합니다.
var errInvalidWrite = errors.New("invalid write result")

// ErrShortBuffer는 읽기 작업에 제공된 버퍼보다 더 긴 버퍼가 필요함을 의미합니다.
var ErrShortBuffer = errors.New("short buffer")

// EOF는 더 이상 읽을 입력값이 없을 때 Read가 반환하는 에러입니다.
// (Read는 EOF를 감싼(wrapping) 에러가 아니라, 반드시 EOF 그 자체를 반환해야 합니다.
// 왜냐하면 호출자들이 == 연산자를 사용해 EOF를 검사하기 때문입니다.)

// 함수들은 입력의 '정상적인 종료 (graceful end)'를 알릴 때만 EOF를 반환해야합니다

// 만약 구조화된 데이터 스트림 도중에 예기치 않게 EOF가 발생한다면,
// 적절한 에러는 [ErrUnexpectedEOF]이거나 더 상세한 내용을 담은 에러여야합니다.

/*
1. 자바는 -1을 주지만, go는 에러를 리턴한다.
2. == 조건으로 EOF를 잡아서 종료하니까 반드시 정확한 EOF를 리턴해야한다.
*/
var EOF = errors.New("EOF")

// ErrUnexpectedEOF는 고정된 크기의 블록이나 데이터 구조를
// 읽는 도중에 EOF를 만났음을 의미합니다.
var ErrUnexpectedEOF = errors.New("unexpected EOF")

/*
자바는  Reader가 데이터를 못읽으면 계속 메모리를 점유하거나 타임아웃이 걸린다.
이 경우일때 발생하는 에러인가..
*/

// 전혀 반환하지 못했을때, 일부 [Reader]의 클라이언트(호출자)가 반환하는 에러
// 이는 보통 고장 난 [Reader] 구현체라는 신호이다
var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Reader 는 기본적인 Read 메서드를 감싸는 인터페이스이다
// Read는 최대 len(p) 바이트 만큼 p로 읽어 들인다
// 읽은 바이트 수 n (0 <= n <= len(p))과 발생한 에러를 반환한다

// Read가 n < len(p)를 반환하더라도, 호출 도중 p의 전체 공간을 임시 공간(scratch space)으로 사용할 수있다
// 만약 데이터가 일부만 있고 len(p)만큼은 없다면, Read는 관례적으로
// 더 기다리기 보다는 현재 사용 가능한 데이터만 반환함 (블로킹 최소)

// Read가 n > 0 바이트를 성공적으로 읽은 뒤 에러나 파일 끝(EOF)를 만난다면
// 일단 읽은 바이트 수를(n) 반환한다
// 이때 (nil이 아닌) 에러를 같은 호출에서 함께 반환할 수도 있고
// 혹은 다음 호출에서 에러(와 n == 0 )를 반환할 수 있음

// 흔한 예로, 입력 스트림의 끝에서 0이 아닌 바이트를 반환한다면
// Reader는 (err == EOF)를 줄 수 있고 (err == nil)을 줄 수도 있음
// 그 다음 Read 호출은 반드시 0, EOF를 반환해야한다

// [핵심] 호출자는 에러(err)를 고려하기전에, 반드시 반환된 n > 0 바이트를 먼저 처리해야한다
// 그래야 일부 바이트 읽은 후 발생한 I/O에러나,
// 허용된 두 가지 EOF 동작 방식을 모두 올바르게 처리할 수 있다

// 만약 len(p) == 0 이면, Read는 항상 n == 0 을 반환해야한다
// EOF 같은 에러 상황이 이미 알려져 있다면 nil이 아닌 에러를 반환할 수 있음

// Read 구현체는 len(p) == 0인 경우를 제외하고는,
// '0 바이트와 nil에러'를 반환하는 것이 권장되지 않는다
// 호출자는 0과 nil 반환을 "아무 일도 일어나지 않음"으로 취급해야하며,
// EOF로 해석해서는 안된다

// 구현체는 파라미터로 받은p를 (호출이 끝난 뒤에도) 붙들고있으면 안된다.

/*
1. 바이트 수를 정해서 읽을 수있게 설계된거같다. 설계시 메모리 최적화에 좋은것같다.
2. 에러가 발생해도 값을 읽을 수 있다
  - 자바는 catch블럭으로 넘어가버리지만 에러가 나더라도 일단 넘어온데이터를 다 쓸 수 있고, 이후 로직도 진행되는거같다.
*/
type Reader interface {
	Read(p []byte) (n int, err error)
}

//writer는 기본적인 write 메서드를 감싸는 인터페이스이다

//writ는 p로부터 len(p)	바이트만큼 기저 데이터 스트림에 쓴다
//p에서 기로된 바이트 수 n (0 <= n <= len(p))과
// 쓰기 작업을 조기에 중단 시킨 에러가 있다면 그 에러를 함께 반환한다.

//만약 n < len(p)를 반환한다면 (다 못썼다면), Write는 반드시 nil이 아닌 에러를 반환한다

//Write는 슬라이스 데이터를 (일시적으로라도) 수행해서는 안된다

// 구현체는 p를 (호출이 끝난 뒤에도)붙들고 있으면 안된다.

/*
Reader는 "len(p) 보다 적게 읽음"이 정상
Writer는 "len(p) 보다 적게 씀"은 무조건 에러 (!= nil)
  - 즉, 반드시 쓰려고 한 만큼 다 써야 정상
*/
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer는 기본적인 Close 메서드를 감싸는 인터페이스이다
// 첫 번째 호출 이후의 Close 동작은 정의되어 있지 않음
// 특정 구현체들은 각자의 동작 방식을 별도로 문서화할 수 있다
/*
자바에서 AutoClosable 을 구현한 녀석들은 close()가 멱등적으로 동작한다.
하지만, go에서는 동작이 정의되어있지 않아 보장되진 않는다.
따라서 개발자가 주의해서 한번만 사용하든 해야할거같다 *defer
*/
type Closer interface {
	Close() error
}

//Seeker는 기본적인 Seek 메서드를 감싸는 인터페이스이다
//Seek은 다음 Read나 Write작업을 위한 오프셋을 설정하며, 그 오프셋은 whence 값에 따라 해석된다
//[SeekStart]는 파일의 시작점을 기준으로,
//[SeekCurrent]는 현재 오프셋(위치)을 기준으로,
//[SeekEnd]는 파일의 끝을 기준으로한다.
// 예를들어, offset = -2는 파일의 끝에서 두 번째 바이트를 가리킨다

//Seek은 파일의 시작점을 기준으로 계산된 "새로운 절대 오프셋"값과,
//발생한 에러를 반환합니다(있다면)

// 양수 오프셋으로 이동하는 것은 허용될 수 있지만,
// 만약 새로운 오프셋이 대상 객체의 크기를 초과한다면
// 그 이후의 I/O 작업들이 어떻게 동작할지는 구현에 따라 다릅니다.
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

//ReadWriter는 기본적인 Read와 Write 메서드를 그룹화한 인터페이스이다
/*
Java에서 interface를 extend한것과 비슷한듯
*/
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser는 기본적인 Read와 Close 메서드를 그룹화한 인터페이스이다
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser는 기본적인 Write와 Close 메서드를 그룹화한 인터페이스이다
type WriteCloser interface {
	Writer
	Closer
}

// ReadWriteCloser 는 기본적인 Read, Write, Close 메서드를 그룹화한 인터페이스이다
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// ReadSeeker는 기본적인 Read와 Seek 메서드를 그룹화한 인터페이스이다
type ReadSeeker interface {
	Reader
	Seeker
}

// ReadSeekCloser는 기본적인 Read, Seek, Close메서드를 그룹화한 인터페이스이다
type ReadSeekCloser interface {
	Reader
	Seeker
	Closer
}

// WriteSeeker는 기본적인 Write와 Seek 메서드를 그룹화한 인터페이스이다
type WriteSeeker interface {
	Writer
	Seeker
}

// ReadWriteSeeker는 기본적인 Read, Write, Seek메서드를 그룹화한 메서드이다
type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// ReaderFrom은 ReadFrom 메서드를 감싸는 인터페이스이다
// ReadFrom은 r로 부터 EOF나 에러를 만날 때 까지 데이터를 읽는다
// 반환값 n은 읽은 바이트 수이다
// 읽는 도중에 발생한 에러 중 EOF를 제외한 모든 에러가 반환된다
// [Copy]함수는 대상객체가 구현했다면 [ReaderFrom]을 우선적으로 사용한다
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo는 WriteTo메서드를 감싸는 인터페이스이다
// WrtierTo는 더 이상 쓸 데이터가 없거나 에러가 발생할 때까지 w에 데이터를 쓴다
// 반환값 n 은 쓴 바이트 수
// 쓰는 도중 발생한 모든 에러가 반환
// [Copy]함수는 소스객체가 구현했다면 WrtierTo를 우선적으로 사용
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

//ReaderAt은 기본적인 ReadAt 메서드를 감싸는 인터페이스이다
//ReadAt은 기저 입력 소스의 오프셋 off위치 부터 시작하여
//len(p) 바이트 만큼을 p로 읽어 들인다
//읽은 바이트 수 n (0 <= n <= len(p))과 발생한 에러를 반환한다

//ReatAt이 n < len(p)를 반환할 때 (요청보다 덜 읽었을 때),
//왜 더 많은 바이트를 반환하지 못했는지 설명하는 nil이 아닌 에러를 반환한다.
//이 점에 있어서 ReadAt은 Read보다 더 엄격하다

//비록 ReadAt이 n <len(p)을 반환하더라도, 호출 도중에 p의 전체 공간을
//임시 공간으로 사용할 수 있다
//만약 데이터가 일부만 있고 len(p) 바이트만큼은 없다면,
//ReadAt은 모든 데이터가 준비되거나 에러가 발생할 때까지 블로킹한다
//이 점이 Read와 다른점이다

// 만약 ReadAt이 탐색 오프셋을 가진 입력 소스로부터 읽는 중이라면,
// ReadAt은 그 기저의 탐색 오프셋에 영향을 주어서도 안되고, 영향을 받아서도 안된다
// ReadAt의 클라이언트들은 동일한 입력 소스에 대해 '병렬로' ReadAt을 호출할 수 있다
// 구현체는 p를 (호출 후에도) 붙들고 있으면 안된다
/*
읽을 위치(off)를 받기 때문에 Thread-safe하다
Read: 일단 받은거 줄 수 있음
ReadAt: 일단 주지 않음. 끝까지 기다렸다가 다 줌 (다 안오면 에러)
*/
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

//WriterAt은 기본적인 WriteAt 메서드를 감싸는 인터페이스이다
//WriteAt은 기저 데이터 스트림의 off 오프셋(위치)에
//p로부터 len(p) 바이트만큼을 쓴다

//p에서 쓴 바이트 수 n (0 <= n <= len(p))과,
//쓰기 작업을 조기에 중단시킨 에러가 있다면 그 에러를 반환한다.

//만약 n < len(p)를 반환하면(다 못썼다면), WriteAt은 반드시 nil이 아닌 에러를 반환해야한다
//만약 탐색 오프셋(seek offset, 커서)이 있는 대상에 WriterAt으로 쓰는 중이라면,
//WriterAt은 그 기저의 탐색 오프셋에 영향을 주어서도 안되고, 영향을 받아서도 안된다.

// WriteAt의 클라이언트들은 쓰기 범위가 겹치지 않는다면,
// 동일한 대상에 대해 '병렬로' WriteAt을 호출할 수 있다
// 구현체는 p를 (호출한 후에도) 붙들고 있으면 안된다.
/*
오프셋이 겹친다면 RaceCondition이 발생할 수 있으므로 주의해야한다
*/
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

//ByteReader는 ReadByte 메서드를 감싸는 인터페이스이다
//ReadByte는 입력으로 부터 다음 바이트를 읽어서 반환하거나,
//발생한 에러를 반환한다

//만약 ReadByte가 에러를 반환한다면, 입력 바이트는 소비되지 않은 상태이며,
//반환된 바이트 값은 정의되지 않음(의미가 없다)

// ReadByte는 '한 번에 한 바이트씩 (byte-at-time)' 처리하기 위한 효율적인 인터페이스를 제공한다.
// ByteReader를 구현하지 않은[Reader]는 bufio.NewReader를 사용해 감싸줌으로 이 메서드를 추가할 수 있다
/*
에러 발생시 byte값은 쓰레기값이 될 수 있으므로 사용x
r := bufio.NewReader(f) -> 버퍼링 처리
*/
type ByteReader interface {
	ReadByte() (byte, error)
}

//ByteScanner는 기본적인 ReadByte 메서드에 UnreadByte 메서드를 추가한 인터페이스이다
//UnReadByte는 바로 다음 ReadByte 호출이 '마지막으로 읽었던 바이트'를 다시 반환하게 만든다

//만약 마지막 연산이 성공적인 ReadByte 호출이 아니라면(예: Write를 했거나, Seek을 했거나 등),
//UnreadByte는 에러를 반환할 수 있고,
//(구현에 따라)마지막으로 읽은 바이트(혹은 그 이전 바이트)를 다시 안 읽은 상태로 돌리거나,
//([Seeker] 인터페이스를 지원하는 구현체라면)현재 오프셋의 1바이트 전으로 탐색(Seek)할 수 있다
/*
보통 직전값만 살릴 수 있음
*/
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

// ByteWriter는 WriteByte 메서드를 감사는 인터페이스이다.
type ByteWriter interface {
	WriterByte(c byte) error
}

//RuneReader는 ReadRune 메서드를 감싸는 인터페이스이다
//RewadRune은 인코딩된 유니코드 문자 하나를 읽고
//그 룬(rune, 문자값)과 그것이 차지하는 바이트 크기를 반환한다
//만약 읽을 수 있는 문자가 없다면 err가 설정됨
/*
rune = int32
문자열을 모두 4바이트로 처리한다
*/
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

//RuneScanner는 기본적인 ReadRune 메서드에 UnreadRune 메서드를 추가한 인터페이스이다
//UnreadRune는 바로 다음 ReadRune호출이 '마지막으로 읽었던 룬(문자)'를 다시 반환한다

//만약 마지막 연산이 성공적인 ReadRune 호출이 아니었다면, UnreadRune은
//에러를 반환할 수 있고
//(구현에 따라)마지막으로 읽은 룬(혹은 그 이전 룬)을 다시 안 읽은 상태로 돌리거나,
// ([Seeker]인터페이스를 지원하는 구현체라면) 현재 오프셋의 '이전 룬의 시작 지점'으로 탐색할 수 있다
/*
UnreadByte -> 무조건 1바이트 이동
UnreadRune -> 숫자, 한글 등에 따라 이동 거리가 다름 (lastRuneSize에 저장)
*/
type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

// StringWriter는 WriteString메서드를 감싸는 인터페이스이다
type StringWriter interface {
	WriteString(s string) (n int, err error)
}

// WriteString은 바이트 슬라이스를 받는 w에 문자열 s의 내용을 쓴다
// 만약 w가 [StringWriter]인터페이스를 구현하고 있다면, [StringWriter.WriterString]이 직접 호출된다
// 그렇지 않으면, [Writer.Write]가 정확히 한 번 호출된다
/*
WriteString 이 내부적으로 최적의 방법을 찾아준다
*/
func WriteString(w Writer, s string) (n int, err error) {
	//1. 타입 단언을 통해 더 효츌적인 메서드가 있는지 찔러본다.
	/*
		바이트 배열을 안쓰고 바로 쓰므로 더 효율적
	*/
	if sw, ok := w.(StringWriter); ok {
		return sw.WriteString(s)
	}
	//2.없다면, 바이트 슬라이스로 변환(메모리 복사 발생)하여 쓴다
	return w.Write([]byte(s))
}

//ReadAtLeast는 r에서 buf로 최소 min 바이트를 읽을 때까지 읽어 들인다
//복사된 바이트 수 n을 반환하며, 만약 min보다 적게 읽혔다면 에러를 반환한다
//바이트를 하나도 읽지 못했을때만 에러가 EOF이다

// 만약 min바이트 보다 적게 읽은 상태에서 EOF가 발생하면 [ErrUnexpectedEOF]를 반환한다
// 만약 min이 buf의 길이보다 크다면 [ErrShortBuffer]를 반환한다
// 반환 시, err == nil인 겨웅에만 n >= min이 성립한다(즉, 성공 시 n >= min)
// 만약 r이 최소 min바이트 이상을 읽은 상태에서 에러를 반환했다면, 그 에러는 무시된다
/*
Named Return Value
 - 메서드 시그니처 부분에, 리턴에 선언된 값(여기선 n, err)는 별도 선언없이 사용
 - 리턴에 변수명을 사용하지 않는다면 일반적으로 쓸 수 있음
 - 리턴에 변수명을 사용하면 선언불가, 할당가능
*/
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}

	for n < min && err == nil {
		var nn int
		//buf[n:] -> 슬라이싱을 통해 "읽어야 할 남은 공간"만 전달
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil //충분히 읽었으면, 마지막에 발생한 에러 (EOF 등)는 무시하고 성공 처리
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF //읽다 말았는데 끊기면 "예기치 못한 EOF"
	}
	return //Named Return Value 사용(n, err 반환)
}

// ReadFull은 r에서 buf로 정확히 len(buf) 바이트만큼 (버퍼 가득)을 읽어들입니다.
// 복사된 바이트 수 n을 반환하며, 만약(버퍼 크기보다) 적게 읽혔다면 에러 반환
// 바이트를 하나도 읽지 못했을때만 에러가 EOF
// 만약 일부는 읽었지만, 전체(버퍼 크기)를 다 읽지 못하면 EOF가 발생한다면 [ErrUnexpectedEOF]를 반환
// 반환 시, err ==nil인 경우만 n == len(buf)가 성립한다
// 즉, 에러가 없으면 버퍼가 꽉 찬다
// 만약 r이 최소 len(buf) 바이트 이상을 읽은 상태에서 에러를 반환한다면, 그에러는 무시
/*
읽기 시작하려는데, 이미 연결이 끊긴 경우
 - 반환값은 0바이트,  err = io.EOF
 	-> 정상 종료로 판단
100바이트 예상하고 읽었지만, 50바이트에서 연결 종료
 - 반환값은 50바이트, err = io.ErrUnexpectedROF
	-> 에러로 판단
*/
func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}
