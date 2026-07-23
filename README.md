# 掼蛋
源项目：`https://github.com/LiUshin/GuanDanInOffice`<br>
使用go、vue3重制<br>
```
guandan/
├── backend/
│   ├── main.go                        # 程序入口，启动 HTTP/WebSocket 服务
│   ├── internal/
│   │   ├── game/
│   │   │   ├── game.go                # 单局游戏核心逻辑 (Game 结构体)
│   │   │   ├── match.go               # 比赛管理 (Match，多局晋级)
│   │   │   ├── room.go                # 房间管理 (Room，玩家集合)
│   │   │   └── bot.go                 # AI 机器人策略
│   │   ├── hub/
│   │   │   ├── hub.go                 # WebSocket Hub（连接管理、消息路由）
│   │   │   └── client.go              # 客户端连接封装
│   │   ├── types/
│   │   │   └── types.go               # 所有数据结构定义 (Card, Hand, GameState, etc.)
│   │   ├── rules/
│   │   │   └── rules.go               # 牌型判定、比较、排序、工具函数
│   │   ├── deck/
│   │   │   └── deck.go                # 牌堆生成、洗牌、属性更新
│   │   └── utils/
│   │       └── utils.go               # 通用辅助函数 (如 ID 生成、日志等)
│   ├── web/                           # 前端构建产物存放目录 (编译时由 Vite 输出)
│   │   └── (静态文件由 embed 嵌入)
│   ├── go.mod
│   ├── go.sum
│   └── Makefile                       # 构建脚本（可选）
├── frontend/
│   ├── public/                        # 静态资源 (favicon, etc.)
│   ├── src/
│   │   ├── main.ts                    # Vue 应用入口
│   │   ├── App.vue                    # 根组件
│   │   ├── router/
│   │   │   └── index.ts               # Vue Router 配置
│   │   ├── store/
│   │   │   ├── index.ts               # Pinia 入口
│   │   │   ├── game.ts                # 游戏状态 store
│   │   │   └── room.ts                # 房间状态 store
│   │   ├── views/
│   │   │   ├── Lobby.vue              # 大厅页面 (加入房间)
│   │   │   └── GameTable.vue          # 游戏桌面 (主游戏界面)
│   │   ├── components/
│   │   │   ├── Card.vue               # 单张扑克牌组件
│   │   │   ├── HandArea.vue           # 手牌区域
│   │   │   ├── PlayerArea.vue         # 其他玩家区域
│   │   │   ├── ChatBox.vue            # 聊天框
│   │   │   ├── SkillCard.vue          # 技能卡按钮
│   │   │   ├── TargetSelector.vue     # 技能目标选择弹窗
│   │   │   └── GameHistory.vue        # 历史记录面板
│   │   ├── composables/
│   │   │   ├── useWebSocket.ts        # WebSocket 连接与消息收发
│   │   │   └── useGameEvents.ts       # 游戏事件监听与分发
│   │   ├── types/
│   │   │   └── index.ts               # 前端类型定义（与后端保持一致）
│   │   └── styles/
│   │       └── index.css              # 全局样式 (Tailwind)
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts                 # Vite 配置 (build 输出到 ../backend/web)
│   ├── tsconfig.json
│   ├── postcss.config.js
│   └── tailwind.config.js
└── README.md
```
