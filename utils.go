package utils

import (
	// "bytes"
	"crypto/rand"
	"encoding/json"
	"strings"

	// "crypto/sha1"
	"fmt"
	"net/http"
	"strconv"

	// "unsafe"

	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

// returns -1 if not found
func FirstIndexOf[T comparable](e T, l []T) int {
	for i, x := range l {
		if x == e {
			return i
		}
	}
	return -1
}

func SprettyPrint(a any) string {
	switch a.(type) {
	case []byte:
		var m map[string]interface{}
		bson.Unmarshal(a.([]byte), &m)
		return SprettyPrint(m)
	case map[string]interface{}, []interface{}, []map[string]interface{}:
		s, _ := json.MarshalIndent(a, "", "    ")
		return string(s)

	}
	panic(fmt.Sprintf("PrettyPrint invalid param type: %T", a))
}

// func FastBytesToString(b []byte) string {
// 	return unsafe.String(unsafe.SliceData(b), len(b))
// }

// func FastStringToBytes(s string) []byte {
// 	return unsafe.Slice(unsafe.StringData(s), len(s))
// }

// func ExtractMediaPlace(mediaId string) string {
// 	i, j := NthIndexOf(2, '-', mediaId), NthIndexOf(3, '-', mediaId)
// 	return mediaId[i+1 : j]
// }

// func NthIndexOf(n int, c byte, s string) int {
// 	var ix int
// 	for ix = strings.IndexByte(s, c); ix > 0; ix = strings.IndexByte(s[ix+1:], c) {
// 		if n--; n > 0 {
// 			return ix
// 		}
// 	}
// 	return ix
// }

func Assert(cond bool, message string, args ...any) {
	if !cond {
		panic(fmt.Sprintf(message, args))
	}
}

func ApplyMiddlewares(server http.Handler,
	handlers ...func(http.Handler) http.Handler) http.Handler {
	for _, h := range handlers {
		server = h(server)
	}
	return server
}

func AddToSet[T comparable](elem T, set []T) ([]T, bool) {
	if Contains(elem, set) {
		return set, false
	} else {
		return append(set, elem), true
	}
}

func AddAllToSet[T comparable](set []T, elems ...T) []T {
	for _, elem := range elems {
		if Contains(elem, set) {
			continue
		} else {
			set = append(set, elem)
		}
	}
	return set
}

func StatusLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		// Serve HTTP request
		lw := &loggingResponseWriter{w, http.StatusOK} // Initialize custom response writer
		next.ServeHTTP(lw, r)
		// Log status code
		t2 := time.Now()
		log.Printf("[%s] %s %s %d %dms\n", r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode, t2.Sub(t1).Milliseconds())
	})
}

// Custom response writer to intercept status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Flush() {
	flusher, ok := lrw.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func HttpLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Panic(err error, errMsg string) {
	if err != nil {
		panic(fmt.Sprintf(errMsg+": %v\n", err))
	}
}

func Fatal(err error, errMsg string) {
	if err != nil {
		log.Fatalf(errMsg+": %v\n", err)
	}
}

func NonFatal(err error, errMsg string) {
	if err != nil {
		log.Printf(errMsg+": %v\n", err)
	}
}

func Backward[E any](s []E) func(func(int, E) bool) {
	return func(yield func(int, E) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(i, s[i]) {
				return
			}
		}
	}
}

// func MakeMongoIds(usernames []string) []string {
// 	return Map(usernames, func(username string) string {
// 		return MakeMongoId(username)
// 	})
// }

// func MakeRawMongoIds(usernames []string) [][12]byte {
// 	return Map(usernames, func(username string) [12]byte {
// 		return MakeRawMongoId(username)
// 	})
// }

// func MakeMongoObjectIds(usernames []string) []primitive.ObjectID {
// 	return Map(usernames, func(username string) primitive.ObjectID {
// 		return MakeMongoObjectId(username)
// 	})
// }

// func MakeMongoId(username string) string {
// 	return MakeMongoObjectId(username).Hex()
// }

// func MakeRawMongoId(username string) [12]byte {
// 	buf := [12]byte{}
// 	hash := sha1.New().Sum([]byte(username))
// 	copy(buf[:], hash)
// 	return buf
// }

// func MakeMongoObjectId(username string) primitive.ObjectID {
// 	buf := [12]byte{}
// 	hash := sha1.New().Sum([]byte(username))
// 	copy(buf[:], hash)
// 	return primitive.ObjectID(buf)
// }

func MakeChatNumUnik(chatNum int) string {
	// a 13 max length, we have max num of 9,999,999,999,999
	// which is 9 trillion+ max chats for a single root
	u := strconv.FormatUint(uint64(chatNum), 10) // base16 (hex)
	padLen := 13 - len(u)
	return strings.Repeat("0", padLen) + u
}

func RandomUnik() string {
	s := base58.Encode(RandomBytes(20))
	s_ := strings.Clone(s[:15])
	if len(s_) != 15 {
		panic("length of random unik string isn't 15")
	}
	return s_
}

func MakeTimeId() string {
	// TODO: implement timeId
	panic("IMPLEMENT TIMEID")
}

func MakeTimestamp() int64 {
	return time.Now().UnixMilli()
}

func MakeTimestampStr() string {
	return strconv.Itoa(int(time.Now().UnixMilli()))
}

func UnixMilli() int64 {
	return time.Now().UnixMilli()
}

func ForEach[T any](l []T, f func(t T)) {
	for _, x := range l {
		f(x)
	}
}

func Map[T any, E any](l []T, f func(e T) E) []E {
	r := make([]E, len(l))
	for i, x := range l {
		r[i] = f(x)
	}
	return r
}

func CopyMap[K comparable, J any](src, dst map[K]J) {
	for k, v := range src {
		dst[k] = v
	}
}

func CopyMap_[K comparable, J any](src map[K]J) map[K]J {
	dst := make(map[K]J, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func CopyMap__[K comparable, J any](src map[K]J) map[K]interface{} {
	dst := make(map[K]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func MaxKey[T comparable](m map[T]int) string {
	mr := func(a [2]interface{}, k T, v int) [2]interface{} {
		if v >= a[1].(int) {
			return [2]interface{}{k, v}
		} else {
			return a
		}
	}

	maxReg := MapReduce(m, [2]interface{}{"", 0}, mr)[0].(string)
	return maxReg
}

func RandomBytes(n int) []byte {
	buf := make([]byte, 0, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func Reduce[T any, E any](l []T, acc E, combine func(a E, b T) E) E {
	for _, x := range l {
		acc = combine(acc, x)
	}
	return acc
}

func MapReduce[K comparable, V any, R any](m map[K]V, acc R, combine func(a R, k K, v V) R) R {
	for k, v := range m {
		acc = combine(acc, k, v)
	}
	return acc
}

func Flatten[T any](m [][]T) []T {
	f := make([]T, 0)
	for _, l := range m {
		f = append(f, l...)
	}
	return f
}

// flattens matrix to array of uniques
func Flattenu[T comparable](m [][]T) []T {
	u := make([]T, 0)
	for _, l := range m {
		for _, x := range l {
			if !Contains(x, u) {
				u = append(u, x)
			}
		}
	}
	return u
}

func Every[T any](l []T, f func(T) bool) bool {
	for _, x := range l {
		if !f(x) {
			return false
		}
	}
	return true
}

func Any[T any](l []T, f func(T) bool) bool {
	for _, x := range l {
		if f(x) {
			return true
		}
	}
	return false
}

func Contains[T comparable](e T, a []T) bool {
	for _, x := range a {
		if x == e {
			return true
		}
	}
	return false
}

func Remove[T comparable](e T, l []T) ([]T, bool) {
	i := FirstIndexOf(e, l)
	if i == -1 {
		return l, false
	}
	return append(l[:i], l[i+1:]...), true
}

func ContainsWhere[T any, K any](e T, a []K, f func(T, K) bool) bool {
	for _, x := range a {
		if f(e, x) {
			return true
		}
	}
	return false
}

func Unique[T comparable](l []T) []T {
	s := make([]T, 0, len(l))
	for _, x := range l {
		if !Contains(x, s) {
			s = append(s, x)
		}
	}
	return s
}

func Filter[T any](l []T, t func(e T) bool) []T {
	f := make([]T, 0, len(l))
	for _, x := range l {
		if t(x) {
			f = append(f, x)
		}
	}
	return f
}
