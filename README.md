# 团队开发项目

使用git进行团队开发开发通常需要遵循下面列出的流程

1. fork团队项目repo到自己的repo中
2. git clone 自己repo中forked的项目，到本地完成开发（对于懂git的人，默认master分支即可，不用重新开branch，除非你能熟练地在使用分支进行更加高级的开发）
3. 使用PR（pull request）向原始repo提交你所做的代码部分。
4. 由项目发起者和组织者进行PR的code review，如果通过code review，则开发者所做的代码将会被merge到主分支中。

上面描述中有很多git相关的术语，如果你不清楚github的使用，直接按照下面的详细步骤进行开发即可。

### fork团队项目repo到自己的repo

点击你要参与开发的项目，比如，该项目

![项目主界面](image/04-staring-tutorial.png)

点击右上角的Fork，然后项目就会被fork到你的名下

### 开发项目

执行如下命令将项目下载到本地进行开发。(你的git_url就是下图那串网址)
```console
git clone your_git_url
```

![自己项目主界面](image/02-git-clone.png)

当你完成自己负责部分项目的开发时，连续使用下面的命令将代码上传到你的github仓库中，第二行中的"Your Message"用于描述代码所做的修改或者所完成的功能，如果你不确定自己完成了什么功能，也可以任意写
```
git add -A
git commit -m "Your Message"
git push origin master
```
*PS: 如果你熟悉你所开发使用的IDE的git相关操作的话，也可以直接通过IDE的git相关按钮完成提交，使用IDE的方式更加简单*
