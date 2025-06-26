## RSS 自动抓取工具

### 配置说明

1. Name：网站名称，会显示在 RSS 订阅源中。
1. URL：目标网站的首页 URL。
2. ItemSelector：文章列表项的 CSS 选择器。
1. TitleSelector：文章标题，相对 ItemSelector 内的选择器。
1. LinkSelector：文章链接，相对 ItemSelector 内的选择器。
1. DescSelector：文章摘要，相对 ItemSelector 内的选择器。
1. DateSelector：文章发布日期，相对 ItemSelector 内的选择器。
1. DateFormat：日期格式，需与网站实际格式一致（参考 Go 的时间格式布局）。

### 使用新增的 RSS 源

http://localhost:8080/rss?site=abc