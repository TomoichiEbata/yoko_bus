/*
サーバ・シミュレータ間で共通に用いる構造体定義等
*/

package routing

// 座標メッセージ
type LocMessage struct {
	ID    int     `json:"id"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	TYPE  string  `json:"type"` // "USER","BIKE"
	POPUP int     `json:"popup"`
	//Address string  `json:"address"`
}

// 座標情報
type LocInfo struct {
	Lng    float64
	Lat    float64
	Source int
}

var LiPoint LocInfo = LocInfo{139.69978753816494, 35.664114318726675, 7565} // 北谷 <<座標ハードコーディング>>
//var LiPoint LocInfo = LocInfo{139.57468655486838, 35.47306323982998, 7565} // 東川島町東 <<座標ハードコーディング>>
