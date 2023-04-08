// https://gtfs.org/realtime/language-bindings/golang/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	sync "sync"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	proto "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

/*
// GetLoc GetLoc
type GetLoc struct {
	ID  int     `json:"id"`
	Lat float32 `json:"lat"`
	Lng float32 `json:"lng"`
	//Address string  `json:"address"`
}
*/

type LocMessage struct {
	ID    int     `json:"id"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	TYPE  string  `json:"type"` // "USER","BIKE"
	POPUP int     `json:"popup"`
	//Address string  `json:"address"`
}

// 構造体の作り方
type unmTbl struct {
	//uniName string // User Name: Example  6ca90e
	uniName int    // new
	objType string // "Bus" or "User"
	simNum  int
	pmNum   int
	lon     float64
	lat     float64
}

var list = make([]unmTbl, 0) // 構造体の動的リスト宣言
//var addr = flag.String("addr", "0.0.0.0:8080", "http service address") // テスト
//var addr = flag.String("addr", "localhost:8080", "http service address") // テスト
//var addr = flag.String("addr", "127.0.0.1:8080", "http service address") // テスト
////var addr = flag.String("addr", "172.26.7.19:8080", "http service address") // テスト
////var addr = flag.String("addr", "54.64.6.92:8080", "http service address") // テスト
var addr = flag.String("addr", "c-anemone.tech:8080", "http service address") // テスト


var mutex sync.Mutex

func main() {

	var wg sync.WaitGroup
	// サーバとのコネクションを1つに統一
	_ = websocket.Upgrader{} // use default options

	flag.Parse()
	log.SetFlags(0)
	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo2"}
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/echo2"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "warn")

	var (
		username = "xx@gmail.com" // 横浜市交通局の市営バスのサイトでは不要のようだからダミーを放り込んでおく
		password = "xx"           // (同上)
	)

	for {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://api.odpt.org/api/v4/gtfs/realtime/YokohamaMunicipalBus_vehicle?acl:consumerKey=f4954c3814b207512d8fe4bf10f79f0dc44050f1654f5781dc94c4991a574bf4", nil)
		if err != nil {
			log.Fatal(err)
		}

		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(body)

		feed := gtfs.FeedMessage{}
		err = proto.Unmarshal(body, &feed)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(feed)

		//for {

		objType := "BUS"
		for _, entity := range feed.Entity {
			//tripUpdate := entity.TripUpdate
			//fmt.Println(entity)

			// データの読み込み
			uniName, _ := strconv.Atoi(*(entity.Vehicle.Vehicle.Id))
			lat := float64(*(entity.Vehicle.Position.Latitude))
			lon := float64(*(entity.Vehicle.Position.Longitude))

			//fmt.Println(uniName, lat, lon)

			flag := 0

			for i := range list {
				if i != 0 && list[i].uniName == uniName { // 同一IDを発見したら
					list[i].lon = lon // 新しい経度情報の更新
					list[i].lat = lat // 新しい緯度情報の更新

					flag = 1
					break
				}
			}

			uniNum := len(list)
			//fmt.Println("---------->uniNum:", uniNum)

			if flag == 0 { // 新しいIDを発見した場合
				wg.Add(1) // goルーチンを実行する関数分だけAddする

				//リストはここで作る
				ut := unmTbl{} // 構造体変数の初期化
				ut.uniName = uniName
				ut.objType = objType
				ut.simNum = uniNum
				ut.lat = lat
				ut.lon = lon

				list = append(list, ut) // 構造体をリストに動的追加
				//fmt.Println("len(list):", len(list))

				// uniNum は、0,1,2,3,4....と増えていく通番、uniNameはデータから送られてくる4ケタのID
				go movingObject(ut, uniNum, uniName, objType, lon, lat, &wg, c)
				flag = 2
			}

		}

		/*
			// 登場しなくなったアイコンを消す為の処理
			for i := range list {

				if list[i].lon < 999.0 || list[i].lat < 999.0 { // 2回目の削除情報は送らない

					fmt.Println("step 1")
					flag2 := 0
					for _, entity := range feed.Entity {
						if list[i].uniName == *(entity.Vehicle.Vehicle.Id) {
							fmt.Println("i:", i, "list[i].uniName:", list[i].uniName, "entity.Vehicle.Vehicle.Id", *(entity.Vehicle.Vehicle.Id))

							flag2 = 1 // 同じ値がでてくれば、これ以上検索する必要はない
							fmt.Println("step 1-2")
							break

						}
						//fmt.Println("step 2")
					}

					if flag2 == 0 { // 上記のfor ループを抜けた → listが多いことになる
						list[i].lon = 999.9 // 削除用の経度情報
						list[i].lat = 999.9 // 削除用の緯度情報
						fmt.Println("Erasing....")
					}
					fmt.Println("step 3")
				}
			}
		*/

		fmt.Println(time.Now())
		time.Sleep(30 * time.Second) // 30秒単位で動きがある様子
	}

	/* unreachable
	// movingObjectに自己破壊メッセージを送信
	// 破壊情報の書き込み中は邪魔させない
	mutex.Lock()
	for i := range list {
		if i != 0 {
			list[i].lon = 999.9 // デタラメな経度情報の更新
			list[i].lat = 999.9 // デタラメな緯度情報の更新
		}
	}
	mutex.Unlock()

	// goルーチンで実行される関数が終了するまで待つ。
	// wg.Wait() // のを止める
	c.Close()
	*/

}

//func movingObject(uniNum int, uniName string, objType string, lon float64, lat float64, wg *sync.WaitGroup, c *websocket.Conn) {
func movingObject(ut unmTbl, uniNum int, uniName int, objType string, lon float64, lat float64, wg *sync.WaitGroup, c *websocket.Conn) {

	//fmt.Printf("start movingObject\n")

	defer wg.Done() // WaitGroupを最後に完了しないといけない。

	//defer c.Close()  // 単一通信だからこれが切れると困る

	// リストを作る前にテストをする
	/*
		fmt.Println("uniNum:", uniNum)
		fmt.Println("uniName:", uniName)
		fmt.Println("objType:", objType)
		fmt.Printf("%f\n", lon)
		fmt.Printf("%f\n", lat)
	*/
	/*
		ut := unmTbl{} // 構造体変数の初期化
		ut.uniName = uniName
		ut.objType = objType
		ut.simNum = uniNum
		ut.lat = lat
		ut.lon = lon
	*/

	gl := new(LocMessage)
	//gl.ID = 0
	gl.ID = -1
	gl.Lat = ut.lat
	gl.Lng = ut.lon
	gl.TYPE = objType
	gl.POPUP = uniName

	mutex.Lock()           // 送受信時にミューテックスロックしないと
	err := c.WriteJSON(gl) // PruneMobile登録用送信
	if err != nil {
		log.Println("write1:", err)
	}

	gl2 := new(LocMessage) // PruneMobile登録確認用受信
	err = c.ReadJSON(gl2)
	if err != nil {
		log.Println("gl2:", err)
	}

	mutex.Unlock()

	ut.pmNum = gl2.ID // PrumeMobileから提供される番号
	//fmt.Println("pmNum = gl2.ID", gl2.ID) // PrumeMobileから提供される番号

	///// ここでlistを作るのは不味い
	/*
		//fmt.Printf("ut.objType=%v\n", ut.objType)
		list = append(list, ut) // 構造体をリストに動的追加
		fmt.Println("len(list):", len(list))
	*/

	// ここからは更新用のループ
	for {
		time.Sleep(time.Millisecond * 100) // 0.1秒休む
		//time.Sleep(time.Second * 10) // 10秒休む

		// 前回との座標に差が認められれば、移動させる
		//diff_lat := float64(list[uniNum].lat - gl.Lat)
		diff_lat := list[uniNum].lat - gl.Lat
		//diff_lon := float64(list[uniNum].lon - gl.Lng)
		diff_lon := list[uniNum].lon - gl.Lng

		if math.Abs(diff_lat) > 0.000000001 || math.Abs(diff_lon) > 0.000000001 {

			//fmt.Print("MOVING!\n")
			gl.Lat = list[uniNum].lat
			gl.Lng = list[uniNum].lon
			gl.ID = gl2.ID

			// 座標の送信

			mutex.Lock()
			err = c.WriteJSON(gl)
			if err != nil {
				log.Println("write2:", err)
			}

			// 応答受信
			gl3 := new(LocMessage)
			err = c.ReadJSON(gl3)
			if err != nil {
				log.Println("gl3:", err)
			}

			mutex.Unlock()

			// 異常値によって、上記でブラウザのオブジェクトを消滅させ、さらに、ここでmovingObjectスレッドも消滅させる
			if list[uniNum].lat > 999.0 || list[uniNum].lon > 999.0 {
				//time.Sleep(10 * time.Second) // 通信が完了するまで待つ時間(▽ざっくり)
				// listから消す必要があるのかもしれないけど、下手にパックすると、順番が怖いので、スレッドを消すだけとする
				println("stop movingObject!")
				return
			}

		}

	}

}
