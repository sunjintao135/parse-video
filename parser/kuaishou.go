package parser

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

type kuaiShou struct{}

func (k kuaiShou) parseShareUrl(shareUrl string) (*VideoParseInfo, error) {
	client := resty.New()
	// disable redirects in the HTTP client, get params before redirects
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, err := client.R().
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7").
		Get(shareUrl)
	// 非 resty.ErrAutoRedirectDisabled 错误时，返回错误
	if !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		return nil, err
	}
	log.Println("shareUrl---------" + shareUrl)
	log.Println("res----------" + res.String())

	// 获取 cookies： did，didv
	cookies := res.RawResponse.Cookies()
	log.Println("cookies开始")
	log.Println(cookies)
	if len(cookies) <= 0 {
		return nil, errors.New("get cookies from share url fail")
	}

	locationRes, err := res.RawResponse.Location()
	if err != nil {
		return nil, err
	}

	log.Println("locationRes----------" + locationRes.String())
	// 分享的中间跳转链接不太一样, 有些是 /fw/long-video , 有些 /fw/photo
	referUri := strings.ReplaceAll(locationRes.String(), "v.m.chenzhongtech.com/fw/long-video", "m.gifshow.com/fw/photo")
	referUri = strings.ReplaceAll(referUri, "v.m.chenzhongtech.com/fw/photo", "m.gifshow.com/fw/photo")
	videoId := strings.ReplaceAll(strings.Trim(locationRes.Path, "/"), "fw/long-video/", "")
	videoId = strings.ReplaceAll(videoId, "fw/photo/", "")
	log.Println("referUri + videoId" + referUri + videoId)
	if len(videoId) <= 0 {
		return nil, errors.New("parse video id from share url fail")
	}

	postData := map[string]interface{}{
		"fid":               "0",
		"shareResourceType": "PHOTO_OTHER",
		"shareChannel":      "share_copylink",
		"kpn":               "KUAISHOU",
		"subBiz":            "BROWSE_SLIDE_PHOTO",
		"env":               "SHARE_VIEWER_ENV_TX_TRICK",
		"h5Domain":          "m.gifshow.com",
		"photoId":           videoId,
		"isLongVideo":       false,
	}
	videoRes, err := client.R().
		SetHeader("Origin", "https://m.gifshow.com").
		SetHeader(HttpHeaderReferer, strings.ReplaceAll(referUri, "m.gifshow.com/fw/photo", "m.gifshow.com/fw/photo")).
		SetHeader(HttpHeaderContentType, "application/json").
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		SetCookies(cookies).
		SetBody(postData).
		Post("https://m.gifshow.com/rest/wd/photo/info?kpn=KUAISHOU&captchaToken=")

	if err != nil {
		return nil, err
	}
	log.Println("log.Println(videoRes.Body()) 开始")
	log.Println(videoRes.Body())
	data := gjson.GetBytes(videoRes.Body(), "photo")
	log.Println(data)
	avatar := data.Get("headUrl").String()
	author := data.Get("userName").String()
	title := data.Get("caption").String()
	videoUrl := data.Get("mainMvUrls.0.url").String()
	cover := data.Get("coverUrls.0.url").String()
	log.Println("title--------------" + title)
	log.Println("videoUrl--------------" + videoUrl)
	log.Println("cover--------------" + cover)

	// 获取图集
	imageCdnHost := data.Get("ext_params.atlas.cdn.0").String()
	imagesObjArr := data.Get("ext_params.atlas.list").Array()
	images := make([]string, 0, len(imagesObjArr))
	if len(imageCdnHost) > 0 && len(imagesObjArr) > 0 {
		for _, imageItem := range imagesObjArr {
			imageUrl := fmt.Sprintf("https://%s/%s", imageCdnHost, imageItem.String())
			images = append(images, imageUrl)
		}
	}

	parseRes := &VideoParseInfo{
		Title:    title,
		VideoUrl: videoUrl,
		CoverUrl: cover,
		Images:   images,
	}
	parseRes.Author.Name = author
	parseRes.Author.Avatar = avatar

	return parseRes, nil
}
