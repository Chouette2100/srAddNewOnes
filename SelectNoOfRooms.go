package main
import (
	"log"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srdblib"
)
/*
指定した順位の範囲 ib 〜 ie にあるルームの数を数える。

	status: 0	ルームの数が ie-ib+1 以上であれば範囲内のルームがすべてデータ取得対象になっている
	status: 1	ルームの数が ie-ib+1 であれば範囲内のルームがすべてデータ取得対象になっている
	status:-1	ランキングのないイベントである。
*/
func SelectNoOfRooms(eventinf *exsrapi.Event_Inf) (
	status int,	//	status: 0 指定した範囲のルーム数は妥当、 1: ルーム数が不足している -1: ランキングがないイベントである。
	err error,
) {
	ib := eventinf.Fromorder
	ie := eventinf.Toorder
	eventid := eventinf.Event_ID
	norooms := 0

	//	データ取得範囲にある対象ルームの数を得る。
	sqls1 := "select count(*) from showroom.points "
	sqls1 += " where eventid = ? and `rank` >= ? and `rank` <= ? "
	sqls1 += " and ts = (select max(ts) from showroom.points where eventid = ? )"
	srdblib.Db.QueryRow(sqls1, eventid, ib, ie, eventid).Scan(&norooms)

	status = 0
	if norooms == 0 {
		//	指定した範囲の順位のルームが存在しない。
		zrooms := 0
		sqls2 := "select count(*) from showroom.points "
		sqls2 += " where eventid = ?  and `rank` = 0 "
		sqls2 += " and ts = (select max(ts) from showroom.points where eventid = ? )"
		srdblib.Db.QueryRow(sqls2, eventid, eventid).Scan(&zrooms)

		if zrooms != 0 {
			//	順位が0のルームが存在するからこのイベントはランキングがないイベントである。
			log.Printf("   *** level event.\n")
			status = -1
		} else {
			status = 1
		}
	} else {
		if norooms < ie-ib+1 {
			status = 1
		}
	}

	return
}
