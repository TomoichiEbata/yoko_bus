// (デバッグ目的で)コンソールからアクセスするための定義。
// Chrome コンソールに、例えば debugs["markers"][123]などとアクセスできる。
let debugs = {};

// onload: index.htmlにおいて最初に呼び出すメインの関数
function onload() {



        var leafletMap = L.map('map').setView([35.66427089602689, 139.69979442662455], 9);
        L.tileLayer("http://{s}.sm.mapstack.stamen.com/(toner-lite,$fff[difference],$fff[@23],$fff[hsl-saturation@20])/{z}/{x}/{y}.png")
            .addTo(leafletMap);


        L.canvasOverlay()
            .drawing(drawingOnCanvas)
            .addTo(leafletMap);










const CENTER_LATLNG = new L.LatLng(35.664114318726675, 139.69978753816494);  // 北谷公園の中心座標
const FES_COORDINATE = [35.664114318726675, 139.69978753816494];  // 北谷公園の表示場所(目的地)

const socket = new WebSocket('wss://127.0.0.1:8080/echo'); // websocketの確立

// socket2に関する処理 → 削除済み5

/* map の表示準備 */
const map = L.map("map", {
    // attributionControl: false,
    zoomControl: false
}).setView(CENTER_LATLNG, 16);

L.tileLayer('https://{s}.tile.osm.org/{z}/{x}/{y}.png', {
    detectRetina: true,
    maxNativeZoom: 18
}).addTo(map);

const leafletView = new PruneClusterForLeaflet(1, 1);  // (120,20)がデフォルト
map.addLayer(leafletView);


/* 北谷公園（fes）のマーカーの作成 */
const fesIcon = L.icon({
    iconUrl: '../images/fes.png',
    iconSize: [60, 36], iconAnchor: [30, 18]
});
// ポップアップが必要であれば、popupAnchor:[0,-60]を付ける。

const fesMarker = L.marker(
    FES_COORDINATE,
    { popup: 'Kitaya', draggable: true, opacity: 1.0, icon: fesIcon }
).addTo(map);

fesMarker.on('dragstart', () => { // マーカーがドラッグしている最中
    fesMarker.setOpacity(0.6); // 透明度0.6に
});
fesMarker.on('dragend', () => { // マーカーが停止した時
    fesMarker.setOpacity(1); // 透明度1.0へ
    console.log(fesMarker);
});


/* socket に関する処理（エージェントの座標を逐次受け取って更新）*/
socket.onopen = function (event) {
}

let markers = [];

// 受信すると、勝手にここに飛んでくる
socket.onmessage = function (event) {
    // データをJSON形式に変更
    let obj = JSON.parse(event.data);

    console.log("233");
    console.log("obj.id:", obj.id);
    console.log("obj.lat:", obj.lat);
    console.log("obj.lng:", obj.lng);
    console.log("obj.type:", obj.type);
    console.log("obj.popup:", obj.popup);

    if (obj.id == -1) {
        let marker = 0
        if (obj.type == "PERSON") {
            marker = new PruneCluster.Marker(obj.lat, obj.lng, {
                popup: "Person " + obj.popup,
                icon: L.icon({
                    iconUrl: '../images/person-icon.png',
                    iconAnchor: [12, 50]
                })
            });
        }
        else if (obj.type == "BIKE") {
            marker = new PruneCluster.Marker(obj.lat, obj.lng, {
                popup: "Bike " + obj.popup,
                icon: L.icon({
                    iconUrl: '../images/bus-icon.png',
                })
            });
        }
        else if (obj.type == "LRT") {
            marker = new PruneCluster.Marker(obj.lat, obj.lng, {
                popup: "LRT " + obj.popup,
                icon: L.icon({
                    iconUrl: '../images/lrt-icon.png',
                    iconAnchor: [34, 13]
                })
            });
        }

        console.log(marker.hashCode);
        markers.push(marker);

        leafletView.RegisterMarker(marker);

        console.log(markers);
        console.log(markers.length);

        obj.id = marker.hashCode;
        //socket.send(marker.hashCode); // テキスト送信
        const jsonObj = JSON.stringify(obj);
        socket.send(jsonObj);
    } else if ((Math.abs(obj.lat) > 90.0) || (Math.abs(obj.lng) > 180.0)) { // 異常な座標が入った場合は、マーカーを消去する
        console.log("Math.abs(obj.lat) > 180.0)")
        for (let index = 0; index < markers.length; ++index) {
            if (obj.id == markers[index].hashCode) {
                console.log(index);
                console.log(obj.id);
                console.log("obj.id == markers[index].hashCode");

                //leafletView.RemoveMarkers(markers[obj.id]);  // これでは消えてくれません
                // 1つのマーカーを消すのに、面倒でも以下の2行が必要
                const deleteList = markers.splice(index, 1);
                leafletView.RemoveMarkers(deleteList);
                break;
            }
        }
        //obj.lat = 91.0;
        //obj.lng = 181.0;
        const jsonObj = JSON.stringify(obj);
        socket.send(jsonObj);
    } else {
        // 位置情報更新
        console.log("else")
        for (let index = 0; index < markers.length; ++index) {
            if (obj.id == markers[index].hashCode) {
                let markerPosition = markers[index].position;
                markerPosition.lat = obj.lat;
                markerPosition.lng = obj.lng;
                break;
            }
        }
        const jsonObj = JSON.stringify(obj);
        socket.send(jsonObj);
    }
}

debugs["markers"] = markers;
debugs["leafletView"] = leafletView;

// サーバを止めると、ここに飛んでくる
socket.onclose = function (event) {
    socket = null;
}


/* 位置情報の更新（１秒毎）*/
window.setInterval(function () {
    leafletView.ProcessView();  // 変更が行われたときに呼び出されれなければならない
}, 1000);

}
