# AGENT.md — mygo湖 项目规范

给参与这个项目的人（以及 AI agent）看的约定。写新代码前先读一遍，评审按这里来。

## 1. 这是什么

mygo湖 是云湖（Yunhu）的桌面客户端：**纯 Go**，窗口用 MyGo 的原生 UI 画，没有
WebView，没有前端构建，产出一个二进制。

- API 文档：<https://yh-api.yyyyt.top/api/v1>
- UI 库：`github.com/egoist/mygo`
- 消息编解码：手写 protobuf wire format，不引 protoc，不用 cgo

## 2. 分层：界面和后台服务分开

三层，依赖只允许往下走：

```
internal/app  ──►  internal/yh  ──►  net/http, coder/websocket
 界面 + 状态          后台服务
```

| 层 | 位置 | 该做的事 | 不该做的事 |
|---|---|---|---|
| 后台服务 | `internal/yh` | HTTP 路由、protobuf 编解码、WebSocket、会话文件 | 不 import `ui`，不画任何东西 |
| 应用层 | `internal/app`（非 `view_*`） | 持有状态、发起请求、把结果搬回界面线程 | 不直接写 HTTP，不写 protobuf |
| 视图层 | `internal/app/view_*.go` | 读状态、画控件 | 不改业务状态、不发网络请求 |

硬规矩：

- `internal/yh` 里出现 `ui.` 就是错的。
- `internal/app` 里出现 `http.` 或手写 protobuf 就是错的——那是 `internal/yh` 的活。
- `internal/yh` 里出现面向用户的文字就是错的——它只回状态码和 error，文案在
  `internal/i18n`（见 §4）。
- 网络调用一律 `go func() { ... a.update(func(){ ... }) }`：在后台线程发请求，
  回到界面线程改状态。视图函数里不允许阻塞。
- 视图里改状态只有一种合法情况：用户交互（点击、输入），且必须落到 `App` 的方法上。

## 3. 目录与文件命名

```
.
├── main.go                 入口、环境变量开关
├── AGENT.md                本文件
├── go.mod
├── internal/
│   ├── yh/                 后台服务层
│   │   ├── api.go          REST 路由
│   │   ├── login.go        手机号登录三步：验证图、短信码、换 token
│   │   ├── ws.go           WebSocket 长连接
│   │   ├── wire.go         protobuf wire format 读写
│   │   ├── types.go        数据结构
│   │   ├── decode.go       状态 / 账号 / 会话列表的解码
│   │   ├── decode_message.go  消息与内容的解码
│   │   ├── decode_ws.go    推送帧与流式片段的解码
│   │   ├── expression.go   收藏表情（JSON 路由）
│   │   ├── profile.go      用户资料（protobuf 路由）
│   │   ├── errors.go       错误种类（只有码，没有文案）
│   │   └── session.go      本机会话文件
│   ├── i18n/               界面文案（按语言分文件，不绘制）
│   │   ├── i18n.go         T / Set / Lang
│   │   ├── zh-CN.json      中文词条
│   │   └── en-US.json      英文词条
│   └── app/                界面
│       ├── app.go          App 状态、窗口生命周期
│       ├── session.go      登录 / 登出
│       ├── live.go         服务端推送
│       ├── convs.go        会话列表状态
│       ├── chat.go         当前会话状态
│       ├── view.go         窗口骨架（view / shell）
│       ├── view_*.go       各个区域的绘制
│       ├── menu.go         弹出菜单（公共控件）
│       ├── controls.go     设置页的公共控件（卡片、行、缩略图选择）
│       ├── switch.go       开关（公共控件，带动画）
│       ├── select.go       下拉选择（公共控件）
│       ├── settings.go     设置窗口的状态、左侧菜单表、打开窗口
│       ├── view_settings.go        设置窗口骨架（左菜单 + 页面）
│       ├── view_settings_*.go      设置窗口的各个页面（通用/通知/存储/快捷键/登录/调色盘/关于）
│       ├── disk_windows.go         磁盘剩余（Windows）
│       ├── disk_other.go           磁盘剩余（别的平台：没有）
│       ├── prefs.go        设置的值（读写配置目录、切换布局、外观、文件夹）
│       ├── stickers.go     表情面板（拉收藏、点着发）
│       ├── usercard.go     用户卡片（点头像开、问 /v1/user/get-user）
│       ├── view_classic.go 经典模式：窄面板
│       ├── chatwin.go      经典模式里单独开的聊天窗口
│       ├── dock.go         贴边隐藏（滑到屏幕边缘、指针回来再滑出）
│       ├── cursor_windows.go       取指针位置（Windows）
│       ├── cursor_other.go         取指针位置（别的平台：没有）
│       ├── notice.go       桌面通知（防抖、点击回会话）
│       ├── preview.go      图片预览窗口（第二个窗口、缩放平移）
│       ├── theme.go        两套配色（深色 / 浅色）
│       ├── icons.go        图标
│       ├── format.go       状态转文字（不绘制）、line 类型
│       ├── images.go       图片缓存
│       ├── markdown.go     markdown 渲染
│       └── html.go         富文本取字
└── shots/                  截图（测试产物）
```

命名规则：

- **`view_<区域>.go` 只画不存。** 新增一块界面就加一个 `view_xxx.go`，别往
  `view.go` 里堆。
- **公共控件自带状态和绘制，不进 `view_*.go`。** `menu.go` 就是一个：`menu` 存内容，
  `showMenu(c, anchor, &m)` 画在任意元素旁边（自动上下翻转、点外面或 Esc 关）。
  新控件照这个来——一个文件、一套 API、谁用都长一样，不要每处各画一遍。- 其余文件按「它管什么状态」命名，一个文件一件事。
- **单个文件超过 300 行就该拆。** 拆分按职责，不按行数硬切。
- 测试文件与被测文件同名加 `_test.go`；界面断言用 `ui.NewTester`。

## 4. i18n

现状：已接入。界面文案全在 `internal/i18n`，绘制代码里没有硬编码的中文。
语言不自己选：每帧从桌面读（`c.Locale()`），所以系统语言一改，窗口就跟着改。

- **应用自己声明会哪几种语言**：`main.go` 里 `mygo.App.SetLanguages("zh-CN", "en-US")`。
  不声明的话 mygo 拿桌面语言去比它自己的词条目录，中文桌面会拿到英文窗口。
- **mygo 自己的字**（菜单、滚动条、无障碍）是另一套词条，每种语言一个包，要 import：
  `_ "github.com/egoist/mygo/locales/zh-Hans"`。

- 所有面向用户的字符串走 `i18n.T`，不在绘制代码里写死中文。
- 词条文件 `internal/i18n/<lang>.json`，key 用 `区域.用途`，例如 `chat.send`、
  `conv.empty`。
- **不做字符串拼接。** 用带占位符的整句：`i18n.T("chat.failed", reason)`，不要在
  外面拼 `"发送失败：" + err`。占位符按出现顺序取参数。
- 数字、时间走 `golang.org/x/text` 的 `message` / `language`：`i18n.T` 里的数字
  参数由 `message.Printer` 写成本地写法（`1,234`），不要自己 `fmt.Sprintf`。
  日期格式也是词条（`time.clock`、`time.date`），不是写死的 layout。
- 布局不能假设文案长度：按钮和标签允许换行或被截断，不要写死宽度。
- 新增文案必须同时给 `zh-CN` 和 `en-US`，缺一个算未完成
  （`TestEveryKeyIsInEveryLanguage` 会红）。

**跨线程的文案存 key，不存句子。** `App` 上要留给后面画的字用 `line`
（`say("conv.count", n)` 或 `words(err.Error())`），画的时候才 `String()`。
否则切换语言后，之前算好的那句话还是旧语言。

**后台服务不写文案。** `internal/yh` 只给状态码（`yh.StateOnline`），
文字由界面层翻（`connState`）。服务器自己返回的错误文本用 `words` 原样显示，
那是它的话，不是我们的词条。

## 5. 消息列表：时间、未读、跳转

- **时间不是每条都写。** 一段连续说话只在开头写一次（`startsStretch`）：第一条、
  隔了 5 分钟以上（`gapLine`）、跨天。跨天按日历判断，不按小时差
  （`dayLine` / `whenLine`），否则凌晨 0:30 会把昨晚 23:00 说成「昨天」以外的日子。
- **未读分两个下标。** `unreadFrom` 是「进来时哪些是新消息」的起点（由会话列表的
  `Unread` 反推，往前翻页时跟着平移），横线画在它上面；`readFrom` 是「读到哪儿」
  的游标，只增不减（`markRead`）。**进会话就是在读**，所以往回翻历史不会把已读的
  又算成未读。
- **跳转按钮数的是 `len(msgs) - readFrom`**（`unreadBelow`）：往下滚就减少，滚到
  底就是 0，按钮消失；点一下回到 `unreadFrom`（`jumpTo` 在列表构建时消费，不能
  在构建之后滚，否则要等下一帧）。
- **往前翻到八成自动续**（`wantMore` / `farBack`）：看到的第一行落在最上面五分之一
  里（`first <= rows/5`）就再要一页。三道闸：
  - `hasMore`：服务器说没有了就不再问；
  - `loadingMore`：上一页还在路上；
  - `loadedAt` + `moreAt`：同一个位置不重复问（一帧问一次会把整段历史抽干），
    两次之间至少 `moreWait`（400ms）。
  **`loadedAt` 必须跟着页走**（`prepend` 里 `+= len(fresh)`）：页插在视图上方，
  列表保位之后「当前看到的第一行」也往下挪了同样多，闸门不跟着挪的话，拖到最顶上
  问过一页之后就再也问不动了——顶上是最小行号，永远小于不了自己。
  同理，**一页没取回来（网络错）要把 `loadedAt` 放回 `1<<30`**：坐在顶上的人没法
  再往上滚，不放手就永远卡住。
- **注意 `farBack` 数的是行，不是像素**：`ui.ListState` 只给可见行号，所以「八成」
  是按条数算的。消息高度差得远（图、长文）时，触发点会跟着偏；想要按像素就得让
  mygo 把滚动偏移和内容高度暴露出来。
- **两边气泡同色**（`bubbleOther`）。谁说的由**位置**区分（自己在右、别人在左带头像和
  名字），不靠颜色——自己那侧没有第二种底色，`palette` 里也没有 `bubbleMine`。
- **图片和表情不加气泡。** 气泡是给文字的，给照片套框就是加相框。表情的图在
  `msg_content.image`，可能是相对路径，要补 `https://chat-img.jwznb.com/`
  （`stickerURL`）。
- **引用是气泡里的一块**（`quoteBlock`）：`msg_content.quote_msg_text` 写成
  「名字: 内容」，按第一个冒号切开，名字一行、内容一行；鼠标在上面时背景变亮并加
  阴影。所以 `messageText` **不再**把 `QuoteText` 当 detail 返回——否则同一句话会
  在气泡里出现两次。
- **时间是一条药丸**（`timeLine`，底色 `qq.timePill`），不是一行灰字：它属于整段
  历史，不属于谁说的话。单条消息的时间是**鼠标停在那一行才出现**的（`row.Hovered()`，
  格式 `time.stamp.full`）。
- **点头像开用户卡片**（`usercard.go`）：`/v1/user/get-user`（protobuf，请求 `id`
  在字段 2）拿名字/头像/备注/签名/性别/地区。卡片画在**窗口的浮层**里，不是用
  `PopoverBase` 挂在行上——列表的行是在布局时建的，那里建的弹层会拿错皮肤，而且同一个
  人每有一条消息就多画一张。
- **点图片开一个独立窗口**（`preview.go`）：`showPicture` 只改状态，窗口在这一帧
  结束之后才建（`a.update` 走的是 postMain；在构建视图时建窗口等于在绘制中途建
  窗口）。图片不参与布局：`viewerFrame` 把图 `Absolute()` 摆到算好的位置上，所以
  缩放和平移只要改 `zoom` / `offX` / `offY`。
- **预览窗口看细节**：Ctrl+滚轮（`InputScroll` 带 `ui.Ctrl`，指数缩放
  `1.0025^-DY`）、触控板捏合（`InputZoom` 的 `Scale`）、按住拖动。缩放以指针为
  中心（`zoomAt`），范围 1/4 到 8 倍；`clampOff` 只允许移到「图片超出窗口的那部分」
  为止，拖不丢。`drawnW` / `drawnH` 由上一帧记下来，`zoomAt` 里跟着一起缩放，
  这样这一帧的夹紧就是按新尺寸算的。
- **新消息弹桌面通知**（`notice.go`）。当前会话**在看着**的时候不弹（`live.go` 直接
  追加了），但窗口被别的窗口挡住或缩在任务栏里时还是弹（`away()`）——那正是通知有用
  的时候。防抖 500ms（`noticeWait`）：每来一条就把定时器重排一次（`armNotice`），
  burst 结束才弹一条；多个会话同时来消息时只弹**最后说话的那个**，正文带
  「{n}条新消息：{text}」。正文是 `preview` + `clip` 的一行预览，不是整条消息。
  通知 id 是 `yh:<chatID>:<chatType>`，点通知回来 `onNoticeClick` 解出会话再
  `openChat`——会话列表还没加载也能打开。`App.notify` 是给测试用的钩子，生产路径走
  `mygo.NewNotification`。

## 6. 登录：密码 / 手机号 / Token

三种方式共用一张卡片，顶上一排 tab 切换（`loginMode` = `modePassword` /
`modePhone` / `modeToken`）。原来是右下角一个「用 Token 登录」的链接，第三种方式塞
不下了才改成 tab。

**手机号登录是三步**（`internal/yh/login.go`）：

1. `POST /v1/user/captcha` 取一张图（`b64s` 是 data URL，`id` 是这张图的身份）；
2. `POST /v1/verification/get-verification-code` 交手机号 + 图上的字 + 图 id，服务端
   才发短信；
3. `POST /v1/user/verification-login` 用手机号 + 短信码换 token，之后和邮箱登录一样
   （`afterLogin`）。

几条规矩：

- **一张图只能读一次。** 发送成功、或者服务端拒了（`APIError`，图被用掉了），都要
  重新取一张（`fetchCaptcha`）；纯网络错误不换，图还在。
- **60 秒才能再发一条**（`smsWait` / `smsLeft`），按钮上直接倒计时（`login.resend`），
  倒计时期间用 `c.After(time.Second)` 每秒要一帧——不要在绘制里等。
- **图不按固定尺寸画。** 接口给的是 180×40，槽位 132×30，按比例塞进去
  （`captchaW` / `captchaH`），拉满就会被压扁。
- 登录成功后把号码存进 session（`Session.Phone`，取自 `u.Phone`）；token 失效回到
  登录页时，只有手机号没有邮箱的账号自动切到手机号那栏，号码已经填好。
- 文案：`login.tab*` / `login.need*` / `login.codeSent` / `err.captcha`。

## 7. 外观

窗口跟随系统，没有自己的开关：`view` 每帧读 `c.Theme().Dark`，取对应的配色，
再把同色系的 `ui.Theme` 交给控件（按钮、输入框、滚动条）。

- 深色背景 `#222222`，浅色背景 `#f5f5f5`；其余颜色是这两级台阶的邻居，
  写在 `theme.go` 的 `darkSkin` / `lightSkin` 里。
- **绘制代码只写 `qq.xxx`，不写颜色字面量。** 需要新颜色就往 `palette` 加字段，
  两套配色都填，别在 `view_*.go` 里 `ui.Hex`。
- `qq` 是当前帧的配色，由 `view` 在每帧开头赋值；它只在绘制线程上读写。

## 8. 设置窗口

设置是**一个窗口**：左边一列分区，右边是选中分区的页面（`settings.go`）。不是每个
分区一个窗口——那会变成每个开关一个窗口。它从侧栏底部的菜单打开（`menu.settings`），
和图片预览一样是第二个窗口，同样用 `mygo.NewWindow` + `TitleBarHidden`：窗口的
最小化/最大化/关闭由桌面画在右上角，视图从顶上第一格开始画，顶部 `c.TitleBar()`
的高度留给它，并且可以拖动（`DragWindow`）。

- **左侧菜单是一张表**（`settingsSections`）：图标 + 词条 key，`apart` 表示另起一组。
  加分区就加一行，不要在绘制里 `if` 出来。菜单的行高、选中底色都在
  `view_settings.go`。
- **页面按分区拆文件**：`view_settings_<分区>.go`，`settingsBody` 里分发。没写页面的
  分区就是空页，这比画一半强。分区下面还能有自己的页（存储空间、更多快捷键）：
  `settings.sub` 记着，标题左边出现返回箭头（`settingsTitle`），值是
  `storageSpacePage` / `hotkeysMorePage`。
- 现在的页面：通用（界面模式/外观/聊天/文件/主面板/浏览）、消息通知（桌面弹窗、托盘
  与消息盒子、窗口抖动、提示音、逐个会话的接收方式）、存储管理（两个目录 + 存储空间
  环形图）、快捷键（两页 + 冲突提醒）、登录设置、超级调色盘、关于。权限设置和账号
  绑定还没有页面（QQ 那两页没截给我）。
- **真的生效的**：界面模式、外观、发送键、双击开窗口、"/"、表情面板、文件目录、
  贴边隐藏（`prefs.Dock`）、通知（类型/详情/好友与群开关/每个会话的 quiet 与 blocked
  都进了 `tellArrival`）、存储空间的数字（真的走目录 + 图片缓存 + 磁盘剩余）。
- **只存值的**：任务栏按钮、关闭行为、内置查看器、默认浏览器、托盘图标样式、窗口抖动、
  提示音（没有播放音频的接口）、开机自启与自动登录（还没调系统接口）、快捷键
  （是应用内的键，不是系统级全局热键）、调色盘主题（只记名字，还没改配色）。
  这些在页面上照常能点，点了会存下来——但别在文档里写成已经生效。
- **页面上的控件是公共的**（`controls.go` / `switch.go` / `select.go`）：卡片、行、
  分区标题、缩略图选择、开关、下拉。谁用都长一样，别在页面里各画一遍。
  - `settingCard` + `settingRow`：卡片管左右内边距，行与行之间的线由后一行画
    （`divider`），所以卡片首尾没有线。
  - `choiceBox`：缩略图 + 下面的字，整块可点（点字也算）。
  - `selectBox`：触发器本身宽 148 高 24，打开时列表**盖住**它（`AttachTo` 左上对齐
    左上），选中的那项打勾；点外面或 Esc 关，由 `PopoverBase` 管。
  - `switchBox`：28×16 的轨道 + 10 的圆钮，用 `AnimateTo` 把「开」当成 0→1 的数，
    钮的位置和轨道颜色都从它算——所以按到一半再按回去是滑回去，不是跳。桌面要求
    减少动效时 `ui` 自己会瞬间到位。
- **设置值存在 `settings` 上**，下拉的值存词条 key（`settings.value.*`）而不是句子，
  理由同 §4：换语言后要跟着变。
- **外观设置那一栏只是记下选择**：窗口外观仍然跟桌面走（§7），自己选一个会是第二个
  找外观的地方。
- **缩略图里画的是「另一个外观的窗口」**，所以它的颜色来自 `shotSkin`
  （`theme.go`），不是 `qq`：白天那张在深色窗口里也得是白的。
- **设置的值在 `prefs` 里**（`prefs.go`），写在配置目录的 `settings.json`，和 session
  挨着。设置窗口读写的就是这个，不是自己的副本：`pickPrefs` 改值 + 存盘 + 让窗口跟上。
  `New()` 只给默认值，**`Start()` 才读盘**——测试里 `New()` 不能去读用户自己的配置。
- **已经接上行为的**：界面模式（换布局 + 改窗口大小）、外观（白天/夜间/跟随系统，
  `a.dark(c)` 在四个窗口入口统一解析，改完所有窗口重画）、发送消息键（Enter /
  Ctrl+Enter，在 composer 里 `HandleInput` 里判）、双击会话开独立窗口（效率模式）、
  "/" 快捷打开（开表情面板）、表情推荐（面板里点一下就发）、文件保存位置（
  `mygo.Dialog.Open` 选目录，菜单「文件」和聊天头上的文件夹图标打开它）。
- **表情面板**（`stickers.go`）：`POST /v1/expression/list` 拿收藏（JSON，390 条那种），
  面板浮在输入框上方，点一张就发。发的时候 `content_type=7`，`msg_content.image` 是
  相对路径、`expression_id`（proto 字段 15，**字符串**）是表情 id——少给 id 服务端会当成
  图片，之后就加不回收藏。图片地址要补 `https://chat-img.jwznb.com/`。
- **还没接的**：发图片/文件/语音（要七牛上传：`/v1/misc/qiniu-token` → 查上传 host →
  上传 → 再发消息，这条链还没写）、通话/屏幕共享/群应用（没有对应 API）、收藏/调色盘/
  聊天记录管理（没有对应 API）、「独立窗口」那种每个会话一个窗口（要多份会话状态）。

## 9. 经典模式与贴边隐藏

**两种布局是同一个窗口的两种排法**（`view.go` 的 `shell` 分发到 `view_classic.go`），
由 `prefs.Mode` 决定，切换时窗口跟着改大小（`applyMode`）：效率模式 1000×720，经典
320×640。**经典模式不多画一套侧栏和列表**：它就是 `rail(c, false)` + `convPanel(c, 0)`
（宽度传 0 表示「把剩下的都给我」），顶上多一条账号块。参数 `account` 决定侧栏要不要
画最上面那个头像——经典模式顶上已经有头像了，画两遍就是一张脸出现两次。
经典模式里**没有聊天区**——点一个会话是开一个独立窗口（`chatwin.go`），窗口
里画的就是效率模式那套 `chatPane` + `infoPanel`：一套代码，两个窗口用。

- **贴边隐藏**（`dock.go`）：鼠标离开面板 700ms 后，面板滑到屏幕上离它最近的一条边（左右两条边，或者**上边**——拖到屏幕顶上的面板就该往上收），只留 3px；
  鼠标回到那条边（上边按窗口宽度、左右按窗口高度算范围，横向放宽 6px）就滑回来。
  - 滑回来是**贴着那条边**出来的，不是回到原来的位置：不然鼠标在边上、面板弹在
    别处，指针一离开面板就又收回去，看起来像「弹了一下就没反应」。
  - 窗口的动画不是 `AnimateTo` 能做的（那是元素）：自己按时间插值，每 12ms 一次
    `SetPosition`，220ms 走完，缓动用 `ui.EaseOut`。
  - **指针位置要自己问**：mygo 只把发生在窗口上的事告诉窗口，鼠标移出去没有事件，
    所以每 120ms 问一次桌面（`cursorPoint`；Windows 用 `GetCursorPos`，别的平台返回
    false，面板就不动）。这也是经典面板空闲时每秒还会画几帧的原因。
  - **滑完必须继续排下一帧**（`dockSlide` 里滑到头之后 `c.After(dockPoll)`）：
    滑完就没人排帧的话，监听循环就断了，面板收起来以后再也弹不出来——只有等别的
    事件重画才醒一次。这是「第一次能弹、第二次没反应」的真正原因。
  - 已经在边上就不再触发第二次滑出——`dock.hidden` 为真且指针不在边上时什么都不做。
- 面板窄是它的意义所在：宽到占满屏幕就没地方可收了。

## 10. 运行内存预算

以 **RSS** 计（任务管理器里看到的那个数）：

| 场景 | 上限 | 现状 |
|---|---|---|
| 平常：聊天、翻会话列表 | 20–80 MB | 实测 67–81 MB（heap 6–18 MB） |
| 中等压力：长会话、大量图片、快速滚动 | < 200 MB | heap 实测 51 MB（1000 条消息 + 图片缓存 47 MB 满载） |
| 极端：几千条消息 + 图片缓存满载 | 450–500 MB | 待压测 |

守住预算的做法：

- **图片一律按显示尺寸向 CDN 要**，绝不拿原图：`?imageView2/1/w/N/h/N`（七牛
  mode 1，等比塞进盒子，不放大）。
  - 头像：显示尺寸 ×2（`avatarPixels`）。
  - 消息图片：原图各边 ÷4（`picturePixels`）。
- 图片缓存按**像素字节数**封顶（`imageBudget`），不是按张数——一张 480px 的
  缩略图和一张 76px 的头像不该算作同样的「一张」。
- 只保留当前会话的消息：切换会话时 `a.msgs = nil`。
- 历史靠翻页（`msgPage`），不预加载。
- 测量：`MYGOHU_MEM=1 ./mygohu.exe`，每 5 秒打一行 heap / sys / 图片缓存字节数。

## 11. 代码风格

- `gofmt` 必须干净，`go vet` 必须干净。
- 注释写**为什么**，不写**是什么**。导出的标识符要有注释。
- 不引 cgo，不引 WebView，不引前端构建链。
- 新依赖要能在 `goproxy.cn` 上拿到，并且要说明为什么标准库不够用。

## 12. 测试

- `go test ./...` 必须全绿，且**默认不联网**。
- 需要真实接口的测试用 `MYGOHU_LIVE=1` 打开，默认 skip。
- 界面改动配一个 `ui.NewTester` 的断言，或写一张 `shots/` 截图。
- 截图：
  - 无窗口快照：`go test ./internal/app -run TestSnapshot`，出图
    `chat.png` / `chat-dark.png` / `chat-en.png` / `chat-new.png`（新消息横线）/
    `chat-unread.png`（跳转未读按钮）/ `chat-info.png` / `menu.png` /
    `viewer.png`（图片预览窗口）/ `login.png` / `login-dark.png` /
    `login-phone.png`（手机号登录：验证图、倒计时）/ `settings.png` /
    `settings-light.png`（设置窗口两套外观）/ `settings-select.png`（下拉展开）/
    `classic.png`（经典窄面板）/ `chat-window.png`（单独开的聊天窗口）/
    `chat-stickers.png`（表情面板）
  - 真窗口截图：`MYGOHU_CAPTURE=shots/x.png ./mygohu.exe`（RDP 下抓屏全黑，
    只有这个办法能看）。设置窗口是第二个窗口：
    `MYGOHU_SETTINGS=1 MYGOHU_CAPTURE_SETTINGS=shots/settings-real.png ./mygohu.exe`
    直接把它开出来并抓图，不用登进聊天再点菜单。
  - **第二个窗口的抓图前要先让它自己画一帧**（`CaptureSettings` 里
    `Invalidate` + 等一下）：抓的是最后一帧的场景，而字形图集是各窗口共用的，
    刚被另一个窗口写过就会抓出别人的字——屏幕上没问题，只有图是坏的。
- 开关这类带动画的东西：`AnimateTo` 的值要能在帧与帧之间断言（按下那一帧钮还在原处，
  `Advance` 之后才过去），见 `settings_test.go`。
- 通知、缩放这类「有副作用」的东西要有钩子才测得动：`App.notify` 收下本该弹到桌面
  的那条通知，定时器、正文、会话 id 就能在测试里断言，不用真弹窗。
- 走到 `afterLogin` 的测试会把 session 写进**用户自己的**配置目录，覆盖真账号：先
  `t.Setenv("APPDATA", t.TempDir())`（Linux 再加 `XDG_CONFIG_HOME`）。设置同理：
  `New()` 不读盘，`Start()` 才读，但 `pickPrefs` 会写，测设置时要 `t.Setenv`。
- 贴边隐藏只能在真窗口上看：它动的是窗口位置，`ui.Tester` 没有窗口。改完用
  `./mygohu.exe` 跑起来，鼠标移开看它收不收、移到边上看出不出来（远程会话里
  可以用 `SetCursorPos` + `mouse_event` 模拟）。

## 13. 云湖 API 备忘

- 服务端 `https://chat-go.jwzhd.com`，WebSocket `wss://chat-ws-go.jwzhd.com/ws`。
- 握手后服务端**不回 login ack**，直接推二进制帧；别等 ack 判断连上没有。
- `POST /v1/msg/list-message` 返回**从新到旧**，入库前要反转。
- 发消息的响应是包了一层的 `StatusResponse`，`code == 1` 才是成功。
- 图片 CDN 有防盗链，必须带 `Referer: https://chat-go.jwzhd.com`，否则 403。
- 消息内容类型：1 文本、2 图片、3 Markdown、4 文件、5 表单、7 表情、8 HTML、
  9 提示、11 语音、13 通话。
- 表情消息的图在 `msg_content.image`（可能是 `expression/xxx.png` 这样的相对路径，
  补 `https://chat-img.jwznb.com/`）；`expression_id` 是 proto 里的 **str**，
  服务端不校验，`sticker_item_id` / `sticker_pack_id` 是表情包那一套。光靠 id
  取不到图，别拿 id 当地址拼。
- 收藏表情列表 `POST /v1/expression/list`（JSON，不是 protobuf）：`data.expression[]`
  里每项 `id`（数字）+ `url`（相对路径）。发的时候 `expression_id` 要**转成字符串**。
- 上传是七牛那一套，还没接：`/v1/misc/qiniu-token`（图片）/ `-audio` / `-video` /
  `qiniu-token2`（文件）拿上传 token，再按 token 里的 ak/bucket 查
  `https://api.qiniu.com/v4/query?ak=&bucket=` 得到上传 host，直传后把返回的 key
  填进 `msg_content.image_url` / `file_key` / `video` / `audio`。
- 连接状态（`yh.State`）只是码：`online` / `retrying` / `offline`，文字在界面层。
- 手机号登录：`/v1/user/captcha` 给 `data.b64s`（`data:image/png;base64,…`，前缀也
  可能写成 `image//png;base64,`，所以只认 `base64,` 之后的部分）+ `data.id`，图是
  180×40 的 PNG；`/v1/verification/get-verification-code` 的字段是
  `mobile` / `code`（图上的字）/ `id`（图的 id）/ `platform`；
  `/v1/user/verification-login` 的字段是 `mobile` / **`captcha`（短信码，不叫 code）**
  / `deviceId` / `platform`。图读错了返回 `code:-1`「图文验证码验证失败」。
