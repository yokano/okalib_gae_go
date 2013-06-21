/**
 * Google App Engine + Go 言語用の汎用ライブラリ
 * package名を自分のアプリ名に合わせて設定してから使用すること
 * @author y.okano
 * @file
 */
package xxxxxx
import (
	"appengine"
	"appengine/urlfetch"
	"appengine/mail"
	"net/http"
	"strings"
	"log"
	"io"
	"math/rand"
	"encoding/binary"
	"encoding/base64"
	"crypto/sha1"
	"time"
	"fmt"
)

/**
 * エラーチェック
 * エラーがあればコンソールに出力する
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {error} err チェックするエラーオブジェクト
 */
func check(c appengine.Context, err error) {
	if err != nil {
		c.Errorf(err.Error())
	}
}

/**
 * スライスから指定された要素を削除して返す
 * 存在しなければ何もしない
 * 削除するのは最初に出現した１つのみ
 * @function
 * @param {[]string} s 対象のスライス
 * @param {string} target 削除する文字列
 * @returns {[]string} 削除済みのスライス
 */
func removeItem(s []string, target string) []string {
	var i int
	var str string
	var result []string
	
	result = make([]string, len(s))
	copy(result, s)
	for i, str = range s {
		if str == target {
			result = append(s[:i], s[i+1:]...)
			break
		}
	}
	
	return result
}

/**
 * 文字列配列の中に指定された文字列が存在するかどうか調べる
 * @function
 * @param {[]string} arr 文字列配列
 * @param {string} target 探す文字列
 * @returns {bool} 存在したらtrue,　それ以外はfalse
 */
func exist(arr []string, target string) bool {
	var i int
	for i = 0; i < len(arr); i++ {
		if arr[i] == target {
			break
		}
	}
	
	result := false
	if i < len(arr) {
		return true
	}
	return result
}

/**
 * 指定されたURLからXMLファイルを受信して返す
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {string} url URL
 * @returns {[]byte} 受信したXMLデータ、取得できなかったら nil を返す
 */
func getXML(c appengine.Context, url string) []byte {
	var client *http.Client
	var response *http.Response
	var err error
	var result []byte
	
	client = urlfetch.Client(c)
	response, err = client.Get(url)
	check(c, err)
	if err != nil {
		log.Printf("URLからファイルを取得出来ませんでした")
		result = nil
	} else {
		result = make([]byte, response.ContentLength)
		_, err = response.Body.Read(result)
		check(c, err)
	}
	
	return result
}

/**
 * スライスの先頭にスライスを挿入する
 * @function
 * @param {[]string} dst 追加されるリスト
 * @param {[]string} src 追加するリスト
 */
func prepend(dst []string, src []string) []string {
	var result []string
	
	result = make([]string, 0)
	result = append(result, src...)
	result = append(result, dst...)
	
	return result
}

/**
 * 文字列を結合する
 * @function
 * @param {string} str 結合する文字列(可変個)
 * @param {string} 結合した文字列
 */
func join(str ...string) string {
	var result string
	var i int
	
	result = str[0]
	for i = 1; i < len(str); i++ {
		result = strings.Join([]string{result, str[i]}, "")
	}
	return result
}

/**
 * リクエストボディ用のリーダー
 * request() で body を送信するために使う
 * @class
 * @member {[]byte} body 本文
 * @member {int} pointer 何バイト目まで読み込んだか表すポインタ
 */
type Reader struct {
	io.Reader
	body []byte
	pointer int
}

/**
 * Reader のインスタンスを作成する
 * @param {string} body 本文
 * @returns {*Reader} 作成したインスタンス
 */
func NewReader(body string) *Reader {
	reader := new(Reader)
	reader.body = []byte(body)
	reader.pointer = 0
	return reader
}


/**
 * 本文を読み出す
 * ２回目以降は前回の続きから読み出せる
 * @method
 * @memberof *Reader
 * @param {[]byte} p 読みだしたデータの保存先
 * @returns {int} 読みだしたバイト数
 * @returns {error} エラー
 */
func (this *Reader) Read(p []byte) (int, error) {
	var l int
	var err error
	if this.pointer + len(p) < len(this.body) {
		l = len(p)
		err = nil
	} else {
		l = len(this.body) - this.pointer
		err = io.EOF
	}
	
	for i := 0; i < l; i++ {
		p[i] = this.body[i + this.pointer]
	}
	
	this.pointer = l + this.pointer
	
	return l, err
}

/**
 * HTTP リクエストを送信してレスポンスを返す
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {string} method POST または GET
 * @param {string} targetUrl 送信先のURL
 * @param {map[string]string} params パラーメタリスト 指定しない場合は nil または空マップ
 * @param {string} body リクエストボディ GET の場合は無視される
 * @param {*http.Response} レスポンス
 */
func request(c appengine.Context, method string, targetUrl string, params map[string]string, body string) *http.Response {
	var request *http.Request
	var err error
	
	// methodのチェック
	if method != "GET" && method != "POST" {
		log.Printf("request(): method must set GET or POST only.")
		return nil
	}
	
	// GET なら URL にクエリ埋め込み
	if method == "GET" && (params != nil || len(params) > 0) {
		paramStrings := make([]string, 0)
		for key, value := range params {
			param := strings.Join([]string{key, value}, "=")
			paramStrings = append(paramStrings, param)
		}
		paramString := ""
		if len(params) == 1 {
			paramString = paramStrings[0]
		} else {
			paramString = strings.Join(paramStrings, "&")
		}
		targetUrl = strings.Join([]string{targetUrl, paramString}, "?")
	}
	
	// リクエスト作成
	if method == "GET" || body == "" {
		request, err = http.NewRequest(method, targetUrl, nil)
	} else {
		request, err = http.NewRequest(method, targetUrl, NewReader(body))
	}
	check(c, err)

	// POST なら Header にパラメータ設定
	if method == "POST" && (params != nil || len(params) > 0) {
		for key, value := range params {
			request.Header.Add(key, value)
		}
	}
	
	// 送受信
	client := urlfetch.Client(c)
	response, err := client.Do(request)
	check(c, err)
	
	return response
}

/**
 * ランダムな文字列を取得する
 * 64bit のランダムデータを Base64 エンコードして記号を抜いたもの
 * @function
 * @returns {string} ランダムな文字列
 */
func getRandomizedString() string {
	r := rand.Int63()
	b := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(b, int64(r))
	e := base64.StdEncoding.EncodeToString(b)
	e = strings.Replace(e, "+", "", -1)
	e = strings.Replace(e, "/", "", -1)
	e = strings.Replace(e, "=", "", -1)
	return e
}

/**
 * SHA-1で暗号化した文字列を返す
 * @function
 * @param {string} 暗号化する文字列
 * @returns {[]byte} 暗号化されたバイト列
 */
func SHA1(input string) []byte {
	hash := sha1.New()
	hash.Write([]byte(input))
	return hash.Sum(nil)
}

/**
 * メールの送信
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {string} sender 送信元アドレス
 * @param {string} to 送信先アドレス
 * @param {string} subject タイトル
 * @param {string} body メッセージ
 */
func sendMail(c appengine.Context, sender string, to string, subject string, body string) {
	message := new(mail.Message)
	message.Sender = sender
	message.To = []string{to}
	message.Subject = subject
	message.Body = body
	
	err := mail.Send(c, message)
	check(c, err)
}

/**
 * クッキーを作成する
 * @function
 * @param {string} name クッキーの名前
 * @param {string} value クッキーの値
 * @param {string} domain 有効ドメイン
 * @param {string} path 有効ディレクトリ
 * @param {int} hour 有効期限（時間）
 */
func NewCookie(name string, value string, domain string, path string, hour int) *http.Cookie {
	duration := time.Hour * time.Duration(hour)
	now := time.Now()
	expire := now.Add(duration)
	
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Domain = domain
	cookie.Path = path
	cookie.Expires = expire
	cookie.RawExpires = expire.Format(time.UnixDate)
	cookie.MaxAge = 60 * 60 * hour
	cookie.Secure = false
	cookie.HttpOnly = true
	cookie.Raw = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	cookie.Unparsed = []string{cookie.Raw}
	
	return cookie
}