/*
シミュレータ用の各種ユーティリティ関数
(DBアクセスが不要なもの)
*/

package routing

import (
	"math"
	"math/rand"
	"time"
)

// 地球の半径
const EARTH_RADIUS = 6378.137

// 指定した範囲の乱数を生成
func Random(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

// 指定した範囲の整数乱数を生成
func RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// 指定した範囲でランダムに決定した"時間"を返す
// Usage: RandomTimeDuration(1, 1000, time.Millisecond)
func RandomTimeDuration(min, max int64, timeSampling time.Duration) time.Duration {
	return time.Duration((rand.Int63n(max-min) + min)) * timeSampling
}

// ラジアン->度
func RadToDeg(a float64) float64 {
	return a / math.Pi * 180.0
}

// 度->ラジアン
func DegToRad(a float64) float64 {
	return a / 180.0 * math.Pi
}

// 地点aから地点bへの距離（返り値1）と角度（返り値2）を計算
func DistanceKm(aLongitude, aLatitude, bLongitude, bLatitude float64) (float64, float64) {
	loRe := DegToRad(bLongitude - aLongitude) // 東西  経度は135度
	laRe := DegToRad(bLatitude - aLatitude)   // 南北  緯度は34度39分

	EWDist := math.Cos(DegToRad(aLatitude)) * EARTH_RADIUS * loRe // 東西距離
	NSDist := EARTH_RADIUS * laRe                                 //南北距離

	distKm := math.Sqrt(math.Pow(NSDist, 2) + math.Pow(EWDist, 2))
	radiusUp := math.Atan2(NSDist, EWDist)

	return distKm, radiusUp
}

// 経度差 [deg]
func DiffLongitude(diffPX, latitude float64) float64 {
	loRe := diffPX / EARTH_RADIUS / math.Cos(DegToRad(latitude)) // 東西
	diffLo := RadToDeg(loRe)                                     // 東西

	return diffLo // 東西
}

// 緯度差 [deg]
func DiffLatitude(diffPY float64) float64 {
	laRe := diffPY / EARTH_RADIUS // 南北
	diff_la := RadToDeg(laRe)     // 南北

	return diff_la // 南北
}

// Sigmoidの逆関数
func InvSigmoid(y float64) float64 {
	// y := rand.Float64() // 0 <= y < 1
	// y = 1 / (1 + exp{-x}) => (1-y)/y = exp{-x} => y/(1-y) = exp x
	return math.Log(y / (1 - y))
}

// Min
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// [min, max]の範囲内に収まるように、値を切り上げ(切り下げ)
func ClipBetween(value, min, max float64) float64 {
	return math.Max(min, math.Min(value, max))
}

// 週末か否かを判定する (通常の「週」 (w=7) だけでなく、一般の周期の「週」を想定)
func IsWeekEnd(year, day int, w int) bool {
	// y年目のd日目 (1≦y≦10、1≦d≦365)
	// y年目のd日目について、((y - 1) * 365 + (d - 1)) を w で割った余りを確認。
	//   ・ 0、又は、w - 1のとき → WEEKEND
	residue := ((year-1)*365 + (day - 1)) % w
	if (residue == 0) || (residue == w-1) {
		return true
	}
	return false
}
