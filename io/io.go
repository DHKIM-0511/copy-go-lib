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
