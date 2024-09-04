/*
Copyright © 2023 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php
*/
package main

import (
	"fmt"
	"log"
	"time"

	//	"math"

	//	"bufio"
	"io"
	"os"

	//	"runtime"

	//	"encoding/json"

	//	"html/template"
	//	"net/http"

	//	"database/sql"
	//	_ "github.com/go-sql-driver/mysql"

	"github.com/go-gorp/gorp"


	//	"github.com/PuerkitoBio/goquery"

	//	svg "github.com/ajstarks/svgo/float"

	//	"github.com/dustin/go-humanize"

	//	scl "UpdateUserInf/ShowroomCGIlib"
	"github.com/Chouette2100/exsrapi"
	//	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

/*
データ取得範囲に新しく加わったルームをデータ取得対象とします。

	イベント参加者テーブル（eventuser）に追加する。
	獲得ポイントテーブル（points）にイベント開始時刻のデータを追加する。

Ver.00AA00 新規作成
Ver.00AB00 ブロックイベント、イベントボックスの親イベントは処理の対象外とする。
Ver.00AC00 ブロックイベントの場合の独自処理を必要最小限にする。
Ver.00AC00aリンクしているsrdblibがv1.1.1となる（ジャンル名の変換規則の変更）
Ver.00AD00 	srdblib.SelectFromEvent()の仕様変更に対応する。
Ver.00AE00 	gorpを導入する。
*/

const Version = "00AE00"

func main() {

	var eventinf *exsrapi.Event_Inf
	//	var roominfolist srdblc.RoomInfoList

	//	ログ出力を設定する
	logfile, err := exsrapi.CreateLogfile(Version, srdblib.Version)
	if err != nil {
		panic("cannnot open logfile: " + err.Error())
	}
	defer logfile.Close()
	//	log.SetOutput(logfile)
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	//	データベースとの接続をオープンする。
	var dbconfig *srdblib.DBConfig
	dbconfig, err = srdblib.OpenDb("DBConfig.yml")
	if err != nil {
		err = fmt.Errorf("srdblib.OpenDb() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer srdblib.Db.Close()

	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	srdblib.Dbmap = &gorp.DbMap{Db: srdblib.Db, Dialect: dial, ExpandSliceArgs: true}
	srdblib.Dbmap.AddTableWithName(srdblib.User{}, "user").SetKeys(false, "Userno")

	log.Printf("********** Dbhost=<%s> Dbname = <%s> Dbuser = <%s> Dbpw = <%s>\n",
		(*dbconfig).DBhost, (*dbconfig).DBname, (*dbconfig).DBuser, (*dbconfig).DBpswd)

	//      cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient("anonymous")
	if err != nil {
		err = fmt.Errorf("CreateNewClient() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	//      すべての処理が終了したらcookiejarを保存する。
	defer jar.Save() //	忘れずに！

	//	現在開催中のイベントのリストを得る
	//	（開始前のものは含まない）
	eventlist, err := srdblib.SelectCurEventList()
	if err != nil {
		log.Printf("SelectCurEventList() returned error. %v", err)
		os.Exit(1)
	}

	for _, event := range eventlist {
		log.Printf(" eventid=[%s] eventname=%s.\n", event.Event_ID, event.Event_name)
		/*	デバッグ時	コメント
			qnow := time.Now().Minute() / 15 // このモジュールが15分に一度起動されることを前提としている。
			hs := time.Since(event.Start_time).Hours()
			he := time.Until(event.End_time).Hours()
			h := math.Min(hs, he)
			if h > 48.0 && qnow != 0 {
				//	開始2日以後かつ終了2日以前の場合は１時間に1回データを取得する。
				continue
			}
			if h > 6.0 && qnow%2 != 0 {
				//	開始6時間以後かつ終了6時間以前の場合は１時間に2回データを取得する。
				continue
			}
				デバッグ時	コメント	*/

		//	イベント情報を取得する。
		eventinf, err = srdblib.SelectFromEvent("event", event.Event_ID)
		if err != nil {
			log.Printf("SelectEventInf() returned error. %v", err)
			continue
		}

		status, err := SelectNoOfRooms(eventinf)
		if err != nil {
			log.Printf("SelectNoOfRooms() returned error. %v", err)
			continue
		}
		if status != 0 {
			//	指定した順位の範囲にデータが取得対象となっていないルームがあるか、レベルイベントである。
			if status == -1 {
				//	レベルイベント
				log.Printf("Event [%s] is level event.\n", event.Event_name)
				if time.Now().Minute()/10%3 != 1 {
					//	10分間隔で起動されることを前提に10分台、40分台で実行する。
					log.Printf("Event [%s] is level event. Not executed.\n", event.Event_name)
					continue
				}
			}

			err = AddNewOnes(client, eventinf)
			if err != nil {
				log.Printf("AddNewOnes() returned error. %v", err)
				continue
			}

		}
	}

}
