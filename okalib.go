/**
 * Google App Engine + Go 言語用の汎用ライブラリ
 * @author y.okano
 * @file
 *
 * package名を自分のアプリ名に合わせて設定してから使用すること
 */
package ******
import (
	"appengine"
	"appengine/urlfetch"
	"net/http"
	"strings"
	"log"
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
