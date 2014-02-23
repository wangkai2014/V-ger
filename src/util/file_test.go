package util

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestExtract(t *testing.T) {
	os.RemoveAll("95edab554826f0e0ebdd0205a3f94dbf")

	resp, err := http.Get("http://res.yyets.com/ftp/2014/0220/95edab554826f0e0ebdd0205a3f94dbf.rar")
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	ioutil.WriteFile("./95edab554826f0e0ebdd0205a3f94dbf.rar", data, 0666)

	Extract("./unar", "./95edab554826f0e0ebdd0205a3f94dbf.rar")

	_, err = os.Stat("95edab554826f0e0ebdd0205a3f94dbf/house.of.cards.2013.s02e11.720p.nf.webrip.dd5.1.x264-ntb/House.of.Cards.2013.S02E11.720p.NF.WEBRip.DD5.1.x264-NTb.繁体.ass")
	if err != nil {
		t.Error(err)
		return
	}
}
