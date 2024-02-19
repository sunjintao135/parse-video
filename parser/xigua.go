package parser

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/PuerkitoBio/goquery"

	"github.com/go-resty/resty/v2"
)

type xiGua struct {
}

func (x xiGua) parseShareUrl(shareUrl string) (*VideoParseInfo, error) {
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, _ := client.R().
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		Get(shareUrl)
	// 这里会返回err, auto redirect is disabled

	locationRes, err := res.RawResponse.Location()
	if err != nil {
		return nil, err
	}

	videoId := strings.ReplaceAll(strings.Trim(locationRes.Path, "/"), "video/", "")
	if len(videoId) <= 0 {
		return nil, errors.New("parse video id from share url fail")
	}

	return x.parseVideoID(videoId)
}

func (x xiGua) parseVideoID(videoId string) (*VideoParseInfo, error) {
	reqUrl := "https://m.ixigua.com/douyin/share/video/" + videoId + "?aweme_type=107&schema_type=1&utm_source=copy&utm_campaign=client_share&utm_medium=android&app=aweme"
	headers := map[string]string{
		HttpHeaderUserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36",
		HttpHeaderCookie:    "MONITOR_WEB_ID=7892c49b-296e-4499-8704-e47c1b150c18; ixigua-a-s=1; ttcid=af99669b6304453480454f150701d5c226; BD_REF=1; __ac_nonce=060d88ff000a75e8d17eb; __ac_signature=_02B4Z6wo00f01kX9ZpgAAIDAKIBBQUIPYT5F2WIAAPG2ad; ttwid=1%7CcIsVF_3vqSIk4XErhPB0H2VaTxT0tdsTMRbMjrJOPN8%7C1624806049%7C08ce7dd6f7d20506a41ba0a331ef96a6505d96731e6ad9f6c8c709f53f227ab1",
	}

	client := resty.New()
	res, err := client.R().
		SetHeaders(headers).
		Get(reqUrl)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return nil, err
	}
	ssrData := doc.Find("#RENDER_DATA").Text()
	ssrJson, err := url.QueryUnescape(ssrData)
	if err != nil {
		return nil, err
	}

	videoData := gjson.Get(ssrJson, "app.videoInfoRes.item_list.0")
	userId := videoData.Get("author.user_id").String()
	userName := videoData.Get("author.nickname").String()
	userAvatar := videoData.Get("author.avatar_thumb.url_list.0").String()
	videoDesc := videoData.Get("desc").String()
	videoAddr := videoData.Get("video.play_addr.url_list.0").String()
	coverUrl := videoData.Get("video.cover.url_list.0").String()

	parseRes := &VideoParseInfo{
		Title:    videoDesc,
		VideoUrl: videoAddr,
		CoverUrl: coverUrl,
	}
	parseRes.Author.Uid = userId
	parseRes.Author.Name = userName
	parseRes.Author.Avatar = userAvatar

	return parseRes, nil
}
