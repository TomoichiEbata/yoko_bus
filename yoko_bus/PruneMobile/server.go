/*
再現シミュレーション用のサーバプログラム

// server.go

// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"m/routing"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // 接続されるクライアント

type key struct {
	*websocket.Conn
	id  int
	att string
}

// 配列宣言
var m1 = make(map[key]int) // このm1は、mmapと異なるクライアント(Webブラウザ単位のMap)

// Ebata: m1保護用のミューテックス
var m1Mutex sync.RWMutex

// ChartData
type ChartData struct {
	UserCnt int `json:"user_cnt"`
	JoinCnt int `json:"join_cnt"`
}

//var addr = flag.String("addr", "127.0.0.1:8080", "http service address") // テスト
//var addr = flag.String("addr", "localhost:8080", "http service address") // テスト
//var addr = flag.String("addr", "192.168.0.8:8080", "http service address") // テスト
var addr = flag.String("addr", ":8080", "http service address") // テスト


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options

var chan2_1 = make(chan routing.LocMessage)
var chan2_2 = make(chan routing.LocMessage)

// chan2_1用のミューテックス
var mutex sync.Mutex

// Enata: map保護用のミューテックス
var mmapMutex sync.RWMutex

//// Ebata: json read write用のmutex
var rwMutex sync.Mutex

// 2次元配列: 変数名は暫定。元々はmmと呼称。
var mmap = map[int]routing.LocMessage{}

func echo2(w http.ResponseWriter, r *http.Request) { // 下からの受けつけ(e.g. Simulator/Hub)
	webConn, err := upgrader.Upgrade(w, r, nil) // cはサーバのコネクション
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer webConn.Close()

	for {
		locMsg := new(routing.LocMessage)

		err := webConn.ReadJSON(&locMsg) // クライアントからのメッセージの受信
		if err != nil {
			log.Println("74: read:", err)
			return // Readロック解除の為、goroutineの強制終了
		}

		mutex.Lock()         // chan2_1を守るミューテックス
		chan2_1 <- *locMsg   // here -> pub
		locMsg2 := <-chan2_1 // pub -> here
		mutex.Unlock()

		err = webConn.WriteJSON(locMsg2) // here -> bike, person
		if err != nil {
			log.Println("write:", err)
			return // Writeロック解除の為、goroutineの強制終了
		}

	}
}

func pub() {

	serialId := 1 // 表示マーカー区別用の通番の初期値

	/*
		redisConn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			panic(err)
		}
		defer redisConn.Close()
	*/

	for { // 番号の対応付け直しを行っている

		//mutex.Lock()        // Ebata:chan2_1を守るミューテックス
		locMsg := <-chan2_1 // echo2(下) -> here
		if locMsg.ID == -1 {
			locMsg.ID = serialId
			serialId += 1 // 表示マーカー区別用の通番のインクリメント
		}
		//mutex.Unlock() // Ebata:chan2_1を守るミューテックス

		mmapMutex.Lock() // map mmap のロック

		/// グローバルマップの作成(ここから)
		_, isThere := mmap[locMsg.ID]

		if isThere && (math.Abs(locMsg.Lat) > 90.0 || math.Abs(locMsg.Lng) > 180.0) { // レコードが存在して、ありえない座標が投入されたら
			//fmt.Println("-----> echo3():enter1")
			delete(mmap, locMsg.ID) // レコードを削除して終了する

		} else if !isThere { // もしレコードが存在しなければ(新しいIDであれば)
			//fmt.Println("-----> echo3():enter2")
			mmap[locMsg.ID] = locMsg // レコードを追加する

		} else { //レコードが存在すれば、要素を書き換える
			//fmt.Println("-----> echo3():enter3")
			mmap[locMsg.ID] = locMsg // レコードの内容を変更する
		}
		/// グローバルマップの作成(ここまで)

		mmapMutex.Unlock() // map mmap のアンロック

		//mutex.Lock() // Ebata:chan2_1を守るミューテックス

		chan2_1 <- locMsg // here -> echo2(下)
		chan2_2 <- locMsg // here -> echo(上)

		/*
			jsonLocMsg, _ := json.Marshal(locMsg)


				///r, err := redis.Int(redisConn.Do("PUBLISH", "channel_1", jsonLocMsg)) //
				_, err := redis.Int(redisConn.Do("PUBLISH", "channel_1", jsonLocMsg)) //
				if err != nil {
					panic(err)
				}
		*/

		//mutex.Unlock() // Ebata:chan2_1を守るミューテックス

		///fmt.Println(r)
	}
}

// UI側(上)とのやり取り
func echo(w http.ResponseWriter, r *http.Request) { // JavaScriptとの通信
	webConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket connection err:", err)
		return
	}
	defer webConn.Close()

	// クライアントを新しく登録(だけ)
	m1Mutex.Lock()
	clients[webConn] = true
	m1Mutex.Unlock()

	/*
		// redisサーバとの接続(subscriber)
		redisConn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			panic(err)
		}
		defer redisConn.Close()
	*/

	// クライアント単位のmapを作る

	// mapの作成

	// map処理を開始する
	/*
		type key struct {
			id  int
			att string
		}
	*/

	/* この宣言はグローバルにしておく必要がある
	type key struct {
		wc  *websocket.Conn
		id  int
		att string
	}

	// 配列宣言
	m1 := make(map[key]int) // このm1は、mmapと異なるクライアント(Webブラウザ単位のMap)
	*/

	/*
		redisPubSubConn := redis.PubSubConn{Conn: redisConn}
		// 購読
		//redisPubSubConn.Subscribe("channel_1", "channel_2", "channel_3")
		//defer redisPubSubConn.Unsubscribe("channel_1", "channel_2", "channel_3")

		redisPubSubConn.Subscribe("channel_1")
		defer redisPubSubConn.Unsubscribe("channel_1")
	*/
	for {
		fmt.Print(time.Now())
		fmt.Print(" 223 ")
		fmt.Println(clients)

		locMsg := new(routing.LocMessage)
		locMsg2 := new(routing.LocMessage)

		*locMsg = <-chan2_2

		var delete_client *websocket.Conn

		for client := range clients { // 全部のクライアントのサーチ

			delete_client = nil

			fmt.Println("231 ")
			// 変数を使って、キーの存在を確認する
			m1Mutex.Lock()
			value, ok := m1[key{client, locMsg.ID, locMsg.TYPE}]
			m1Mutex.Unlock()
			fmt.Println("236")
			/////0423 if math.Abs(locMsg.Lat) > 90.0 || math.Abs(locMsg.Lng) > 180.0 { // ありえない座標が投入されたら
			if ok && (math.Abs(locMsg.Lat) > 90.0 || math.Abs(locMsg.Lng) > 180.0) { // レコードが存在して、ありえない座標が投入されたら
				fmt.Println("enter 1")
				fmt.Println("240")
				tmpId := locMsg.ID /// 0423
				locMsg.ID = value  // mapから見つけた値を使って、

				fmt.Println("1:locMsg:", locMsg)

				rwMutex.Lock() ////Ebata

				err = client.WriteJSON(&locMsg) // 送って
				if err != nil {
					delete_client = client

				}
				err = client.ReadJSON(&locMsg2) // 戻して
				if err != nil {
					delete_client = client
				}

				rwMutex.Unlock() ////Ebata

				m1Mutex.Lock()
				delete(m1, key{delete_client, tmpId, locMsg.TYPE}) // レコードを削除して終了する 0423
				m1Mutex.Unlock()

				fmt.Println("1:locMsg2:", locMsg2)

			} else if !ok { // もしレコードが存在しなければ(新しいIDであれば)
				fmt.Println("enter 2")

				tmpId := locMsg.ID
				locMsg.ID = -1 // 空番号 これでJavaScriptの方に

				fmt.Println("2:locMsg:", locMsg)

				rwMutex.Lock() ////Ebata

				err = client.WriteJSON(&locMsg) // 送って
				if err != nil {
					delete_client = client
				}
				err = client.ReadJSON(&locMsg2) // 戻してもらって
				if err != nil {
					delete_client = client
				}

				rwMutex.Unlock() ////Ebata

				fmt.Println("2:locMsg2:", locMsg2)

				pm_id := locMsg2.ID // JavaScriptから与えられたIDで
				//fmt.Println("id:", id, ", pm_id:", pm_id)

				//time.Sleep(time.Second * 1)
				time.Sleep(time.Millisecond * 10)

				m1Mutex.Lock()
				m1[key{client, tmpId, locMsg.TYPE}] = pm_id // レコードを追加する
				m1Mutex.Unlock()

			} else { //レコードが存在すれば、その値を使ってアイコンを動かす

				//fmt.Println("enter 3")

				locMsg.ID = value // mapから見つけた値を使って、
				// このバグの原因はJavaScript側のsendとrecvのタイミングのズレだった
				//fmt.Println("3:locMsg:", locMsg)

				rwMutex.Lock()                  ////Ebata
				err = client.WriteJSON(&locMsg) // アイコンを動かす
				if err != nil {
					delete_client = client
				}
				client.ReadJSON(&locMsg2)
				if err != nil {
					delete_client = client
				}
				rwMutex.Unlock() ////Ebata
				//fmt.Println("3:locMsg2:", locMsg2)
			}

		}

		/// ループを出てからリストから削除する

		if delete_client != nil {
			fmt.Println("delete_client")
			delete_client.Close()
			delete(clients, delete_client)
		}

	}

}

func echo3(w http.ResponseWriter, r *http.Request) {

	fmt.Println("             Echo3() is starting..........")

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn2, err := upgrader.Upgrade(w, r, nil) //conn2でwebsocketを作成
	if err != nil {
		log.Println("websocket connection err:", err)
		return
	}
	defer conn2.Close()

	for {

		///////////////////////////////////////////////////////////
		chart := new(ChartData)
		joinCnt := 0

		mmapMutex.Lock() // map mmap のロック

		chart.UserCnt = len(mmap) //テーブルエントリの数

		for _, v := range mmap {
			dis, _ := routing.DistanceKm(v.Lng, v.Lat, 139.69978753816494, 35.664114318726675) // 北谷公園

			fmt.Println("dis:", dis)
			if dis < 0.10 { //100メートルに入ったらカウント
				joinCnt += 1
			}
		}

		mmapMutex.Unlock() // map mmap のアンロック

		chart.JoinCnt = joinCnt // rand.Intn(20) // ここで乱数を発生されて、javascriptで受信させる

		err := conn2.WriteJSON(&chart)
		if err != nil {
			log.Println("WriteJSON:", err)
			break
		}
		fmt.Println("echo3:", chart)
		time.Sleep(time.Second * 1) // こいつがガンでした(ブロードキャストの取り逃がし)
	}

}

func main() {

	flag.Parse()
	log.SetFlags(0)

	log.Println(routing.LiPoint)
	go pub()

	// アクセスされたURLから /static 部分を取り除いてハンドリングする
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/echo3", echo3)                                         // echo3関数を登録 (サーバとして必要)
	http.HandleFunc("/echo2", echo2)                                         // echo2関数を登録 (サーバとして必要)
	http.HandleFunc("/echo", echo)                                           // echo関数を登録 (サーバとして必要)
	log.Fatal(http.ListenAndServeTLS(*addr, "./fullchain.pem", "./privkey.pem", nil)) // localhost:8080で起動をセット
}
