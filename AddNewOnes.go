/*!
Copyright © 2023 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php
*/
package main
import (
	"fmt"
	"log"
	"strings"

	"net/http"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

/*
指定した範囲にあってデータ取得の対象となっていないルームをデータ取得の対象として設定する。
*/
func AddNewOnes(
	client *http.Client,
	eventinf *exsrapi.Event_Inf,
) (
	err error,
) {

	var roomlistinf *srapi.RoomListInf

	if strings.Contains(eventinf.Event_ID, "?") {
		//	ブロックイベント
		roomlistinf, err = exsrapi.GetRoominfFromEventOfBR(
			client,
			eventinf.Event_ID,	//	Event_url_key (string)
			eventinf.Fromorder,
			eventinf.Toorder,
		)
		if err != nil {
			err = fmt.Errorf("GetRoominfFromEventOfBR() returned error. %w", err)
			return
		}
	} else {
		//	ブロックイベント以外のイベント
		roomlistinf, err = srapi.GetRoominfFromEventByApi(
			client,
			eventinf.I_Event_ID,	//	Event_id (int)
			eventinf.Fromorder,
			eventinf.Toorder,
		)
		if err != nil {
			err = fmt.Errorf("srapi.GetRoominfFromEventByApi() returned error. %w", err)
			return
		}
	}

	for i, room := range roomlistinf.RoomList {
		log.Printf("room: %d %s", room.Room_id, room.Room_name)
		err = srdblib.InsertNewOnes(client, eventinf.Fromorder+i, *eventinf, room)
		if err != nil {
			//	err = fmt.Errorf("InsertIntoEventUser() returned error. %w", err)
			log.Printf("InsertIntoEventUser() returned error. %v", err)
			continue
		}
	}

	return
}
