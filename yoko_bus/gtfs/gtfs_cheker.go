// https://gtfs.org/realtime/language-bindings/golang/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	proto "github.com/golang/protobuf/proto"
)

func main() {

	os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "warn")

	var (
		username = "xx@gmail.com" // 横浜市交通局の市営バスのサイトでは不要のようだからダミーを放り込んでおく
		password = "xx"           // (同上)
	)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.odpt.org/api/v4/gtfs/realtime/YokohamaMunicipalBus_vehicle?acl:consumerKey=f4954c3814b207512d8fe4bf10f79f0dc44050f1654f5781dc94c4991a574bf3", nil)
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

	fmt.Println(body)

	feed := gtfs.FeedMessage{}
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(feed)

	for _, entity := range feed.Entity {
		//tripUpdate := entity.TripUpdate
		//fmt.Println(entity)

		a := entity.Vehicle.Vehicle.Id
		latitude := entity.Vehicle.Position.Latitude
		longitude := entity.Vehicle.Position.Longitude

		fmt.Println(*a, *latitude, *longitude)
		//trip := tripUpdate.Trip
		//tripId := trip.TripId
		//fmt.Printf("Trip ID: %s\n", *tripId)
	}

}
