# Elasticsearch 镜像：在官方镜像基础上安装 IK 中文分词插件（分析器 ik_max_word / ik_smart）
# 与 docker-compose 中的 es 服务版本保持一致；如插件地址失效，可换用其它 IK 构建。
FROM docker.elastic.co/elasticsearch/elasticsearch:8.15.2

RUN bin/elasticsearch-plugin install --batch https://get.infini.cloud/elasticsearch/analysis-ik/8.15.2
