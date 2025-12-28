package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liuzl/gocc"
)

type Tv struct {
	Channel   []Channel   `xml:"channel"`
	Programme []Programme `xml:"programme"`
}

type Channel struct {
	ID          string      `xml:"id,attr"`
	DisplayName DisplayName `xml:"display-name"`
}

type DisplayName struct {
	Text string `xml:",chardata"`
}

type Programme struct {
	Channel string `xml:"channel,attr"`
	Start   string `xml:"start,attr"`
	Stop    string `xml:"stop,attr"`
	Title   Title  `xml:"title"`
	Desc    Desc   `xml:"desc"`
}

type Title struct {
	Text string `xml:",chardata"`
}

type Desc struct {
	Text string `xml:",chardata"`
}

var (
	cache       sync.Map
	cacheExpiry = 60 * time.Second
	fetchURL    = "https://cdn.jsdmirror.com/gh/dfdg881/myEPG@master/output/epg.xml"
	userAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	epglists    = []string{"CCTV-1","CCTV-2","CCTV-3","CCTV-4","CCTV-5","CCTV-5+","CCTV-6","CCTV-7","CCTV-8","CCTV-9","CCTV-10","CCTV-11","CCTV-12","CCTV-13","CCTV-14","CCTV-15","CCTV-16","CCTV16-4K","CCTV-17","CCTV-4K","CCTV-8K","CCTV1","CCTV10","CCTV11","CCTV12","CCTV13","CCTV14","CCTV15","CCTV16","CCTV17","CCTV2","CCTV3","CCTV4","CCTV4K","CCTV4欧洲","CCTV4美洲","CCTV5","CCTV5+","CCTV5PLUS","CCTV6","CCTV7","CCTV8","CCTV8K","CCTV9","CETV1","CETV2","CETV4","CETV5","CGTN英语","CGTN纪录","CGTN俄语","CGTN法语","CGTN西语","CGTN阿语","中国教育1台","中国教育2台","中国教育4台","文化精品","央视台球","风云音乐","第一剧场","风云剧场","怀旧剧场","女性时尚","高尔夫网球","风云足球","电视指南","世界地理","兵器科技","广东卫视","浙江卫视","湖南卫视","北京卫视","湖北卫视","黑龙江卫视","安徽卫视","重庆卫视","东方卫视","甘肃卫视","广西卫视","贵州卫视","海南卫视","河北卫视","河南卫视","吉林卫视","江苏卫视","江西卫视","辽宁卫视","内蒙古卫视","宁夏卫视","青海卫视","山东卫视","山西卫视","陕西卫视","四川卫视","深圳卫视","三沙卫视","天津卫视","西藏卫视","新疆卫视","云南卫视","康巴卫视","兵团卫视","大湾区卫视","广东民生","动漫秀场","乐游","中国天气","都市剧场","法治天地","东方财经","金色学堂","环球奇观","生态环境","山东教育","纪实科教","纯享4K","金鹰纪实","快乐垂钓","先锋乒羽","茶频道","纪实人文","欢笑剧场","生活时尚","福建文体","福建新闻","福建电视剧","福建经济","福建综合","福建乡村振兴","福建电视剧","福建旅游","东南卫视","海峡卫视","厦门卫视","厦门一套","厦门二套","厦门三套","FZTV1","FZTV3","三明公共","三明新闻综合","云霄综合","宁化电视一套","将乐综合","建宁综合","德化新闻综合","新罗电视一套","晋江电视台","永安综合","永泰综合","泰宁新闻","漳州新闻综合","漳浦综合","石狮综合","霞浦综合","龙岩公共","龙岩新闻综合","云霄综合","建宁综合","漳州新闻","龙岩公共","龙岩综合","重温经典","翡翠台","明珠台","凤凰中文","凤凰资讯","凤凰香港","TVB Plus","无线新闻","RTHK31","RTHK32","RTHK33","RTHK34"}
	converter   *gocc.OpenCC
)

func init() {
	var err error
	converter, err = gocc.New("t2s")
	if err != nil {
		fmt.Println("Error initializing converter:", err)
	}
}

func fetchEPGData() {
	for {
		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}

		req.Header.Set("User-Agent", userAgent)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error: unable to fetch XML data")
			continue
		}

		xmlData, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}

		var tv Tv
		err = xml.Unmarshal(xmlData, &tv)
		if err != nil {
			fmt.Println("Error unmarshalling XML:", err)
			continue
		}

		cache.Store("epg", tv)
		fmt.Println("EPG data updated")

		time.AfterFunc(cacheExpiry, fetchEPGData)
		break
	}
}

func formatDateTime(timeStr string) (string, string) {
	if strings.Contains(timeStr, "-") {
		return timeStr, ""
	}

	if len(timeStr) < 8 {
		return "", ""
	}

	year := timeStr[:4]
	month := timeStr[4:6]
	day := timeStr[6:8]
	date := fmt.Sprintf("%s-%s-%s", year, month, day)

	var time string
	if len(timeStr) >= 12 {
		hour := timeStr[8:10]
		minute := timeStr[10:12]
		time = fmt.Sprintf("%s:%s", hour, minute)
	}

	return date, time
}

func getCurrentDateInBeijing() string {
	TimeLocation, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		TimeLocation = time.FixedZone("CST", 8*60*60)
	}
	currentTime := time.Now().In(TimeLocation)
	return currentTime.Format("2006-01-02")
}

func generateDefaultEPG() []map[string]string {
	epgData := make([]map[string]string, 24)
	for hour := 0; hour < 24; hour++ {
		startTime := fmt.Sprintf("%02d:00", hour)
		endTime := fmt.Sprintf("%02d:00", (hour+1)%24)
		epgData[hour] = map[string]string{
			"start": startTime,
			"end":   endTime,
			"title": "精彩节目-暂未提供节目预告信息",
			"desc":  "",
		}
	}
	return epgData
}

func sanitizeChannelName(channel string) string {
	tid := strings.ToUpper(channel)
	re := regexp.MustCompile(`\[.*?\]|[0-9\.]+M|[0-9]{3,4}[pP]|[0-9\.]+FPS`)
	tid = re.ReplaceAllString(tid, "")
	tid = strings.TrimSpace(tid)
	re = regexp.MustCompile(`超清|高清$|蓝光|频道$|标清|FHD|HD$|HEVC|HDR|4K|-|\s+`)
	tid = re.ReplaceAllString(tid, "")
	tid = strings.TrimSpace(tid)

	if strings.Contains(tid, "CCTV") && !strings.Contains(tid, "CCTV4K") {
		re := regexp.MustCompile(`CCTV[0-9+]{1,2}[48]?K?`)
		matches := re.FindStringSubmatch(tid)
		if len(matches) > 0 {
			tid = strings.Replace(matches[0], "4K", "", -1)
		} else {
			re = regexp.MustCompile(`CCTV[^0-9]+`)
			matches = re.FindStringSubmatch(tid)
			if len(matches) > 0 {
				tid = strings.Replace(matches[0], "CCTV", "", -1)
			}
		}
	} else {
		tid = strings.Replace(tid, "BTV", "北京", -1)
	}
	return tid
}

func getMatchedChannel(query string, tv Tv, date string) string {
	normalizedQuery := sanitizeChannelName(query)
	simplifiedQuery, err := converter.Convert(normalizedQuery)
	if err != nil {
		simplifiedQuery = normalizedQuery
	}

	priorityMatch := ""
	secondaryMatch := ""
	matched := ""

	for _, epg := range epglists {
		upperEPG := strings.ToUpper(epg)
		if strings.HasPrefix(upperEPG, normalizedQuery) || strings.HasPrefix(upperEPG, simplifiedQuery) {
			if hasEPGData(epg, tv, date) {
				return epg
			}
		} else if strings.Contains(upperEPG, normalizedQuery) || strings.Contains(upperEPG, simplifiedQuery) {
			if isChinesePrefix(epg) {
				if priorityMatch == "" {
					priorityMatch = epg
				}
			} else {
				if secondaryMatch == "" {
					secondaryMatch = epg
				}
			}
		}
	}

	if priorityMatch != "" && hasEPGData(priorityMatch, tv, date) {
		return priorityMatch
	}
	if secondaryMatch != "" && hasEPGData(secondaryMatch, tv, date) {
		return secondaryMatch
	}

	for i := len(normalizedQuery); i > 0; i-- {
		subQuery := normalizedQuery[:i]
		for _, epg := range epglists {
			upperEPG := strings.ToUpper(epg)
			if strings.HasPrefix(upperEPG, subQuery) {
				if hasEPGData(epg, tv, date) {
					if len(epg) > len(matched) {
						matched = epg
					}
				}
			}
		}
	}

	if matched != "" {
		return matched
	}

	return "未知频道"
}

func isChinesePrefix(s string) bool {
	re := regexp.MustCompile(`^[\p{Han}]`)
	return re.MatchString(s)
}

func hasEPGData(channel string, tv Tv, date string) bool {
	for _, programme := range tv.Programme {
		if strings.Contains(strings.ToUpper(programme.Channel), strings.ToUpper(channel)) && strings.HasPrefix(programme.Start, strings.ReplaceAll(date, "-", "")) {
			if programme.Title.Text != "" {
				return true
			}
		}
	}
	return false
}

func handleEPG(c *gin.Context) {
	channel := strings.ToUpper(c.DefaultQuery("ch", "CCTV1"))
	dateParam := c.DefaultQuery("date", getCurrentDateInBeijing())
	date, _ := formatDateTime(dateParam)
	epgInterface, exists := cache.Load("epg")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "EPG data not available"})
		return
	}
	tv := epgInterface.(Tv)

	channel = getMatchedChannel(channel, tv, date)

	var epgData []map[string]string
	for _, programme := range tv.Programme {
		if programme.Channel == channel && strings.HasPrefix(programme.Start, strings.ReplaceAll(date, "-", "")) {
			_, startTime := formatDateTime(programme.Start)
			_, endTime := formatDateTime(programme.Stop)
			epgData = append(epgData, map[string]string{
				"start": startTime,
				"end":   endTime,
				"title": programme.Title.Text,
				"desc":  programme.Desc.Text,
			})
		}
	}

	if len(epgData) == 0 {
		epgData = generateDefaultEPG()
	}

	response := map[string]any{
		"date":         date,
		"channel_name": channel,
		"epg_data":     epgData,
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	go fetchEPGData()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/json", handleEPG)
	r.Run(":27100")
}
