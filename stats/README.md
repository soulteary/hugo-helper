# 统计文章发布状况

统计以年月日进行存放的内容存放的 Hugo 网站文章数据，统计数据将保存在 `report/stats.json`。

```bash
docker pull soulteary/hugo-go-stats

docker run --rm -it -v `pwd`/articles:/docs soulteary/hugo-go-stats hugo-go-stats /docs
```
