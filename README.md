#### youdaonote-backup是用Go语言编写的有道云笔记备份命令行软件
软件只是为了自用快速写的, 功能很简单, 所以不适用于每个人的需求
##### 1. 功能
- 下载有道云笔记中的通用格式笔记, 如md等通用格式直接下载
- 将有道云笔记的默认格式即我们平时新建空白文档的格式下载为html格式, 相当于一个折中方案, 但是只是比下载word格式好看一点点, 诸如插入的代码这些显示是有问题的
- 下载文档中插入的图片和附件, 修改文档中的引用路径, 使在本地也可查看备份文档中的图片, 采用了相对路径, 把备份上传云盘后再下载到随机目录文档中的图片仍可以查看。附件只显示一个附件的名称, 需要我们去相应下载目录查看

##### 2. 使用
- linux, mac, windows的可执行文件已经编译好了放在了exe目录下, 下载对应文件即可使用, linux与win的可执行文件只是作者在mac上交叉编译的, 如有问题请反馈
- 使用, 以mac为例:
>  ./note_backup_max_amd64 -d ./back -c "网页中复制的cookie, 注意双引号加上"

> -d 指定备份目录, 如例子中为在当前可执行程序的目录下创建back目录用于备份  
> 注：图片目录在指定目录的picture目录下, 如按示例图片与附件存放在  ./back/picture 中

> -c 指定从登录后的有道云笔记网页中复制的cookie
> 有道云cookie不知如何获取可以利用搜索引擎查询

> 其它参数可以查看代码

##### 3. 可能出错的使用点: 
- 从浏览器中复制的cookie中已经含有了双引号, 此时需要去除cookie中的双引号，cookie的格式应该为：key1=value1; key2=value2
- 下载错了可执行文件, 请观察可执行文件名后下载正确的可执行文件

##### 4. 未完成的工作:
- note格式文件转换为html后, 还是含有网易自定义的标签修饰等, 无法还原在app中的观感
- 遇到部分文件总是下载错误, 目前未定位到原因, 如果是部分文件下载失败, 请确保网络正常, 如果网络正常, 则应该就是作者所说的这个现象