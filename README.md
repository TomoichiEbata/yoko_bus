# yoko_bus

公共交通オープンデータセンター 開発者サイト(https://developer.odpt.org/)の
横浜市交通局 バス関連リアルタイム情報 / Bus... バス関連リアルタイム情報 / Bus...
https://ckan.odpt.org/dataset/b_bus_gtfs_rt-yokohamamunicipal
の、

```
バス関連リアルタイム情報 / Bus realtime information
URL: https://api.odpt.org/api/v4/gtfs/realtime/YokohamaMunicipalBus_vehicle?acl:consumerKey=[発行されたアクセストークン/YOUR_ACCESS_TOKEN]

横浜市交通局の市営バスのバス関連リアルタイム情報を提供します。 / Bus realtime information of Transportation Bureau, City of Yokohama

各車両のリアルタイムの現在地情報 (VehiclePosition) を提供します。
```
を使って、バス関連のリアルタイム情報を取得して、地図上にバスの位置を表示するプログラムです。

[発行されたアクセストークン/YOUR_ACCESS_TOKEN]は、自分で取得する必要があります。
(プログラムのトークンはダミーです)

# 現状
Amazon Lightsail上で動いています。


# 参考メモ
「Protocol Buffersって何？ 」から、「公共交通オープンデータ」を攻略する
https://wp.kobore.net/江端さんの技術メモ/post-9594/
