#!/bin/bash

# 获取所有tag为<none>的镜像ID
none_images=$(docker images --filter "dangling=true" -q)

# 判断是否有<none>的镜像存在
if [ -z "$none_images" ]; then
  echo "No <none> images to remove."
else
  # 删除所有<none>的镜像
  echo "Removing <none> images..."
  docker rmi $none_images
  echo "All <none> images removed."
fi