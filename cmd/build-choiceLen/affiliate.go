package main

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

// Public affiliate URL formally issued by Rakuten and supplied by the user.
// This is a published link, not an API credential. Preserve its query exactly.
const blueairAffiliateURL = "https://hb.afl.rakuten.co.jp/ichiba/5841d747.2dc49921.5841d748.b537efd7/_RTLink149141?pc=https%3A%2F%2Fitem.rakuten.co.jp%2Fblueair%2Fb-111521%2F&link_type=text&ut=eyJwYWdlIjoiaXRlbSIsInR5cGUiOiJ0ZXh0Iiwic2l6ZSI6IjI0MHgyNDAiLCJuYW0iOjEsIm5hbXAiOiJyaWdodCIsImNvbSI6MSwiY29tcCI6ImRvd24iLCJwcmljZSI6MCwiYm9yIjoxLCJjb2wiOjEsImJidG4iOjEsInByb2QiOjAsImFtcCI6ZmFsc2V9"

func enableAffiliate(article string) (string, error) {
	u, err := url.Parse(blueairAffiliateURL)
	if err != nil || u.Scheme != "https" || u.Host != "hb.afl.rakuten.co.jp" || u.Query().Get("pc") != "https://item.rakuten.co.jp/blueair/b-111521/" {
		return "", fmt.Errorf("invalid issued affiliate destination")
	}
	const marker = "<h2>商品情報への導線（未設定）</h2>"
	if strings.Count(article, marker) != 1 {
		return "", fmt.Errorf("affiliate placement marker missing or duplicated")
	}
	start := strings.Index(article, marker)
	end := strings.Index(article[start:], "</main>")
	if end < 0 {
		return "", fmt.Errorf("article closing tag missing")
	}
	footer := `<h2>商品情報への導線</h2><p>購入前チェックで説明書の条件を確認した後に、代表商品の型番・価格・販売条件を確認してください。価格・クーポン・在庫は変動します。</p><p>広告・アフィリエイトリンク：<a rel="sponsored" href="` + html.EscapeString(blueairAffiliateURL) + `">Blueair Blue Mini Max（111521）の販売条件を楽天市場で確認する</a></p>`
	article = article[:start] + footer + article[start+end:]
	return strings.Replace(article, "現在、実アフィリエイトリンクは未設定です。", "楽天市場への広告・アフィリエイトリンクを記事末尾に掲載しています。", 1), nil
}
