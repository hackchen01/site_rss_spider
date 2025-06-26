## RSS 自动抓取工具

### 配置说明

Name：网站名称，会显示在 RSS 订阅源中。
URL：目标网站的首页 URL。
TitleSelector：文章标题的 CSS 选择器。
LinkSelector：文章链接的 CSS 选择器。
DescSelector：文章摘要的 CSS 选择器。
DateSelector：文章发布日期的 CSS 选择器。
DateFormat：日期格式，需与网站实际格式一致（参考 Go 的时间格式布局）。

### 使用新增的 RSS 源

http://localhost:8080/rss?site=abc